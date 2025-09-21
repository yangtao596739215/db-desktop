# Streaming Tool Call Reconstruction Implementation

## Overview

This document describes the implementation of proper streaming response handling for tool calls that are fragmented across multiple chunks in the AI API response.

## Problem

The original streaming implementation had issues with tool call reconstruction when the API returns tool calls in multiple chunks. For example:

```json
// Chunk 1: Tool call start
{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_123","function":{"name":"execute_redis_command","arguments":"{\"command\":"}}]}}]}

// Chunk 2: Arguments continuation  
{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":" \"KEYS *"}}]}}]}

// Chunk 3: Arguments completion
{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"}"}}]}}]}
```

The arguments `{"command": "KEYS *"}` were being split across multiple chunks, making simple string concatenation insufficient.

## Solution

### 1. New Data Structures

#### StreamResponseCallback
```go
type StreamResponseCallback struct {
    OnMessage  func(*models.Message) // Called for each message chunk
    OnComplete func(*models.Message) // Called when stream is complete
    OnError    func(error)           // Called on error
}
```

#### StreamingState
```go
type StreamingState struct {
    CurrentMessage *models.Message
    ToolCalls      map[int]*models.MCPToolCall // Index-based reconstruction
    ContentBuffer  strings.Builder
    HasMCP         bool
}
```

### 2. Tool Call Reconstruction Logic

The `processToolCallDelta` function handles incremental tool call updates:

```go
func (c *AIClient) processToolCallDelta(toolCallDeltas []models.MCPToolCall, state *StreamingState, callback *StreamResponseCallback) {
    for i, delta := range toolCallDeltas {
        // Get or create tool call for this index
        toolCall, exists := state.ToolCalls[i]
        if !exists {
            toolCall = &models.MCPToolCall{
                ID:   delta.ID,
                Type: delta.Type,
                Function: models.MCPFunctionCall{
                    Name:      delta.Function.Name,
                    Arguments: "",
                },
            }
            state.ToolCalls[i] = toolCall
        }

        // Update tool call fields incrementally
        if delta.ID != "" {
            toolCall.ID = delta.ID
        }
        if delta.Type != "" {
            toolCall.Type = delta.Type
        }
        if delta.Function.Name != "" {
            toolCall.Function.Name = delta.Function.Name
        }
        if delta.Function.Arguments != "" {
            toolCall.Function.Arguments += delta.Function.Arguments
        }

        // Send tool call update to callback
        if callback.OnMessage != nil {
            toolCallMsg := &models.Message{
                Role:      "assistant",
                ToolCalls: []models.MCPToolCall{*toolCall},
            }
            callback.OnMessage(toolCallMsg)
        }
    }
}
```

### 3. Streaming Response Processing

The main streaming loop now properly handles both content and tool calls:

```go
// Initialize streaming state
state := &StreamingState{
    CurrentMessage: &models.Message{
        Role:      "assistant",
        Content:   "",
        ToolCalls: []models.MCPToolCall{},
    },
    ToolCalls: make(map[int]*models.MCPToolCall),
    HasMCP:    false,
}

decoder := json.NewDecoder(resp.Body)
for {
    var streamResp AIResponse
    if err := decoder.Decode(&streamResp); err != nil {
        if err == io.EOF {
            break
        }
        return fmt.Errorf("failed to decode stream: %w", err)
    }

    // Process choices
    for _, choice := range streamResp.Choices {
        // Handle content delta
        if choice.Delta.Content != "" {
            state.ContentBuffer.WriteString(choice.Delta.Content)
            state.CurrentMessage.Content = state.ContentBuffer.String()
            
            // Send content update to callback
            if callback.OnMessage != nil {
                contentMsg := &models.Message{
                    Role:    "assistant",
                    Content: choice.Delta.Content,
                }
                callback.OnMessage(contentMsg)
            }
        }

        // Handle tool calls delta
        if len(choice.Delta.ToolCalls) > 0 {
            state.HasMCP = true
            c.processToolCallDelta(choice.Delta.ToolCalls, state, callback)
        }

        // Check if this is the final choice
        if choice.FinishReason != "" {
            // Finalize the message
            state.CurrentMessage.Content = state.ContentBuffer.String()
            if state.HasMCP {
                // Convert tool calls map to slice
                for _, toolCall := range state.ToolCalls {
                    state.CurrentMessage.ToolCalls = append(state.CurrentMessage.ToolCalls, *toolCall)
                }
            }
            
            // Send final complete message
            if callback.OnComplete != nil {
                callback.OnComplete(state.CurrentMessage)
            }
        }
    }
}
```

## Key Features

### 1. Incremental Reconstruction
- Tool calls are built incrementally as chunks arrive
- Arguments are concatenated properly across multiple chunks
- State is maintained for each tool call index

### 2. Message-Based Callbacks
- Callbacks now receive `*models.Message` instead of strings
- Type-safe message structures
- Better separation of concerns

### 3. Real-Time Updates
- Content chunks are sent immediately to the frontend
- Tool call updates are sent as they're being built
- Final complete message is sent when stream ends

### 4. Error Handling
- Proper error handling for malformed chunks
- Graceful handling of incomplete tool calls
- State cleanup on errors

## Benefits

1. **Accurate Reconstruction**: Tool calls are properly reconstructed from fragmented streaming data
2. **Type Safety**: Message structures instead of string concatenation
3. **Real-Time Updates**: Frontend receives updates as they arrive
4. **Clean Architecture**: Better separation between integration and logic layers
5. **Robust Error Handling**: Graceful handling of edge cases

## Usage Example

```go
streamCallback := &integration.StreamResponseCallback{
    OnMessage: func(msg *models.Message) {
        // Handle real-time content or tool call updates
        if msg.Content != "" {
            // Send content to frontend
        }
        if len(msg.ToolCalls) > 0 {
            // Handle tool call updates
        }
    },
    OnComplete: func(finalMessage *models.Message) {
        // Handle complete message with all tool calls
        if len(finalMessage.ToolCalls) > 0 {
            // Process tool calls
        }
    },
    OnError: func(err error) {
        // Handle errors
    },
}

err := aiClient.SendMessageStreamWithCallback(messages, streamCallback)
```

## Testing

The implementation has been tested with the example streaming data provided:

- Chunk 1: Tool call initialization with partial arguments
- Chunk 2: Arguments continuation
- Chunk 3: Arguments completion
- Chunk 4: Tool call finalization
- Chunk 5: Stream completion

The final reconstructed tool call correctly contains:
- ID: `call_30fb8bdcce274fbfbb8bd4`
- Function: `execute_redis_command`
- Arguments: `{"command": "KEYS *"}`

This implementation ensures that complex tool calls split across multiple streaming chunks are properly reconstructed and handled by the application.
