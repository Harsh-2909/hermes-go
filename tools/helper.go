package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

// CreateToolFromMethod creates a Tool from a method of a toolkit instance.
//
// TODO: Changes to be made:
// - Refactor this function to use go/types package instead of go/ast.
// - Make this more readable and maintainable.
// - Add support for more types.
func CreateToolFromMethod(toolkit interface{}, methodName string) (Tool, error) {
	// Get the method via reflection
	method, ok := reflect.TypeOf(toolkit).MethodByName(methodName)
	if !ok {
		return Tool{}, fmt.Errorf("method %s not found", methodName)
	}

	methodType := method.Type

	// Validate method signature: must have ctx as first param after receiver
	if methodType.NumIn() < 2 || methodType.In(1) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return Tool{}, fmt.Errorf("method must have context.Context as first parameter after receiver")
	}

	// Check return values: either one value or (value, error)
	var hasError bool
	if methodType.NumOut() == 1 {
		hasError = false
	} else if methodType.NumOut() == 2 && methodType.Out(1) == reflect.TypeOf((*error)(nil)).Elem() {
		hasError = true
	} else {
		return Tool{}, fmt.Errorf("method must return one value or (value, error)")
	}

	// Get parameter types (excluding receiver and ctx)
	paramTypes := make([]reflect.Type, methodType.NumIn()-2)
	for i := 2; i < methodType.NumIn(); i++ {
		paramTypes[i-2] = methodType.In(i)
	}

	// Get package path and type name from the toolkit
	pkgPath := reflect.TypeOf(toolkit).Elem().PkgPath()
	typeName := reflect.TypeOf(toolkit).Elem().Name()

	// Find the source directory using go/build
	bpkg, err := build.Import(pkgPath, "", build.FindOnly)
	if err != nil {
		return Tool{}, fmt.Errorf("failed to find package %s: %v", pkgPath, err)
	}
	srcDir := bpkg.Dir

	// Parse the package directory to get the AST
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, srcDir, nil, parser.ParseComments)
	if err != nil {
		return Tool{}, fmt.Errorf("failed to parse package %s: %v", pkgPath, err)
	}

	// Assume the first package (typically one package per directory)
	// TODO: ast.Package is deprecated. Migrate to go/types package.
	var astPkg *ast.Package
	for _, p := range pkgs {
		astPkg = p
		break
	}
	if astPkg == nil {
		return Tool{}, fmt.Errorf("no package found in %s", srcDir)
	}

	// Find the method declaration
	var file string
	var line int
	for _, f := range astPkg.Files {
		for _, decl := range f.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Recv != nil {
				if len(fd.Recv.List) == 1 {
					recvType := fd.Recv.List[0].Type
					if star, ok := recvType.(*ast.StarExpr); ok {
						if ident, ok := star.X.(*ast.Ident); ok && ident.Name == typeName {
							if fd.Name.Name == methodName {
								pos := fset.Position(fd.Pos())
								file = pos.Filename
								line = pos.Line
								break
							}
						}
					}
				}
			}
		}
		if file != "" {
			break
		}
	}
	if file == "" {
		return Tool{}, fmt.Errorf("method %s not found on type %s", methodName, typeName)
	}

	// Parse the source file
	fset = token.NewFileSet()
	pkgs, err = parser.ParseDir(fset, filepath.Dir(file), nil, parser.ParseComments)
	if err != nil {
		return Tool{}, fmt.Errorf("failed to parse source file: %v", err)
	}

	var astFile *ast.File
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			if fset.Position(f.Pos()).Filename == file {
				astFile = f
				break
			}
		}
		if astFile != nil {
			break
		}
	}
	if astFile == nil {
		return Tool{}, fmt.Errorf("source file not found")
	}

	// Find the method declaration
	var funcDecl *ast.FuncDecl
	for _, decl := range astFile.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && fset.Position(fd.Pos()).Line == line {
			funcDecl = fd
			break
		}
	}
	if funcDecl == nil || funcDecl.Doc == nil {
		return Tool{}, fmt.Errorf("method %s has no doc comments", methodName)
	}

	// Parse doc comments
	doc := funcDecl.Doc.Text()
	lines := strings.Split(doc, "\n")
	description := strings.TrimSpace(lines[0]) // First line is the description

	paramDescs := make(map[string]string)
	required := make([]string, 0)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)

		// Checking for @param or @params prefix. Either is supported
		var prefix string
		if strings.HasPrefix(line, "@param ") {
			prefix = "@param "
		} else if strings.HasPrefix(line, "@params ") {
			prefix = "@params "
		}
		if prefix != "" {
			paramLine := strings.TrimPrefix(line, prefix)
			isOptional := false

			// Checks for optional parameter
			if strings.HasPrefix(paramLine, "[optional] ") {
				isOptional = true
				paramLine = strings.TrimPrefix(paramLine, "[optional] ")
			}
			parts := strings.SplitN(paramLine, ":", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				desc := strings.TrimSpace(parts[1])
				paramDescs[name] = desc
				if !isOptional {
					required = append(required, name)
				}
			}
		}
	}

	// Get parameter names from AST (skip receiver and ctx)
	paramNames := make([]string, 0, len(paramTypes))
	for _, field := range funcDecl.Type.Params.List[1:] { // Skip ctx
		for _, name := range field.Names {
			paramNames = append(paramNames, name.Name)
		}
	}
	if len(paramNames) != len(paramTypes) {
		return Tool{}, fmt.Errorf("parameter count mismatch")
	}

	// Build JSON schema parameters
	properties := make(map[string]interface{})
	for i, name := range paramNames {
		schemaType, ok := goTypeToJSONSchemaType(paramTypes[i])
		if !ok {
			return Tool{}, fmt.Errorf("unsupported parameter type: %v", paramTypes[i])
		}
		properties[name] = map[string]interface{}{
			"type":        schemaType,
			"description": paramDescs[name],
		}
	}
	parameters := map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}

	// Create the Execute function
	execute := func(ctx context.Context, args string) (string, error) {
		var argMap map[string]interface{}
		if err := json.Unmarshal([]byte(args), &argMap); err != nil {
			return "", fmt.Errorf("failed to unmarshal args: %v", err)
		}

		// Build argument list
		argValues := []reflect.Value{reflect.ValueOf(toolkit), reflect.ValueOf(ctx)}
		for i, name := range paramNames {
			val, ok := argMap[name]
			if !ok {
				// check if the parameter is optional
				if slices.Contains(required, name) {
					return "", fmt.Errorf("missing required parameter: %s", name)
				}
				// Use default value for missing optional parameters.
				switch paramTypes[i].Kind() {
				case reflect.String:
					argValues = append(argValues, reflect.ValueOf(""))
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					argValues = append(argValues, reflect.ValueOf(0))
				case reflect.Float32, reflect.Float64:
					argValues = append(argValues, reflect.ValueOf(0.0))
				case reflect.Bool:
					argValues = append(argValues, reflect.ValueOf(false))
				case reflect.Slice:
					argValues = append(argValues, reflect.ValueOf([]interface{}{}))
				case reflect.Map:
					argValues = append(argValues, reflect.ValueOf(make(map[string]interface{})))
				default:
					return "", fmt.Errorf("optional Parameter %s with type %s not supported", name, paramTypes[i])
				}
			} else {
				converted, err := convertJSONValueToGoType(val, paramTypes[i])
				if err != nil {
					return "", fmt.Errorf("type conversion failed for %s: %v", name, err)
				}
				argValues = append(argValues, reflect.ValueOf(converted))
			}
		}

		// Call the method
		results := method.Func.Call(argValues)
		if hasError {
			if err, ok := results[1].Interface().(error); ok && err != nil {
				return "", err
			}
			result := results[0].Interface()
			jsonResult, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("failed to marshal result: %v", err)
			}
			return string(jsonResult), nil
		}
		result := results[0].Interface()
		jsonResult, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %v", err)
		}
		return string(jsonResult), nil
	}

	return Tool{
		Name:        methodName,
		Description: description,
		Parameters:  parameters,
		Execute:     execute,
	}, nil
}

func goTypeToJSONSchemaType(t reflect.Type) (string, bool) {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer", true
	case reflect.Float32, reflect.Float64:
		return "number", true
	case reflect.String:
		return "string", true
	case reflect.Bool:
		return "boolean", true
	case reflect.Slice:
		return "array", true
	case reflect.Map:
		return "object", true
	default:
		return "", false
	}
}

func convertJSONValueToGoType(val interface{}, t reflect.Type) (interface{}, error) {
	switch t.Kind() {
	case reflect.Int:
		if f, ok := val.(float64); ok {
			return int(f), nil
		}
	case reflect.Float64:
		if f, ok := val.(float64); ok {
			return f, nil
		}
	case reflect.String:
		if s, ok := val.(string); ok {
			return s, nil
		}
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			return b, nil
		}
	case reflect.Slice:
		if arr, ok := val.([]interface{}); ok {
			slice := reflect.MakeSlice(t, len(arr), len(arr))
			for i, v := range arr {
				elem, err := convertJSONValueToGoType(v, t.Elem())
				if err != nil {
					return nil, err
				}
				slice.Index(i).Set(reflect.ValueOf(elem))
			}
			return slice.Interface(), nil
		}
	case reflect.Map:
		if m, ok := val.(map[string]interface{}); ok {
			mapType := reflect.MapOf(reflect.TypeOf(""), t.Elem())
			newMap := reflect.MakeMap(mapType)
			for k, v := range m {
				elem, err := convertJSONValueToGoType(v, t.Elem())
				if err != nil {
					return nil, err
				}
				newMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(elem))
			}
			return newMap.Interface(), nil
		}
	default:
		return nil, fmt.Errorf("unsupported type: %v", t)
	}
	return nil, fmt.Errorf("cannot convert %v to %v", val, t)
}
