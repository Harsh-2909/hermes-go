# Tool Call

## Steps

- Agent will accept the tools as an array of tools `[]Tools` when an agent is created.
- When Agent is Initialized `Init()`, we will parse the Tools to any specified format (or not?), and pass them to the models so that for every future run of that model, the tools part will be populated in the request and we won't have to pass tools from agent to models for each run.
- In the model, we will create the function to parse the tool call format from our `Tool` to the format specified by the model (like `OpenAI`)
- Non Streaming:
	- The response of the model will be captured in `ModelResponse` struct. It needs to save the tool call details as well in `ToolCall` field.
	- The `ToolCall` field will have the `ToolCallID` which will be used to map the tool call request to the tool call response.
	- Run a for loop on the tool calls, call all the tools, get the response and save the response to `agent.Message` with the `ToolCallID` (mapping between model tool call request to our tool call response).
	- Keep running the loop until the response received does not have tool calls, then return the response.
- Steaming:
	- The OpenAI API responds with delta on the args needed for function calls. 
	- When streaming and tool call is present, we need to capture the delta, create a single `ModelResponse` object with `ToolCall`, then send those to `Agent.RunStream`.
	- In `Agent.RunStream`, we will keep running the loop until the response received does not have tool calls, then return the response.
	- We return all the chunks with Role: "assistant" back to the user so that he can only see the final response.
	- When streaming and tool call is absent, we can handle it as it is.

## Reference

- Check the OpenAI docs on [Function Calling](https://platform.openai.com/docs/guides/function-calling) and [Create Chat Completions](https://platform.openai.com/docs/api-reference/chat/create) 
- Agno codebase for inspiration on interface.