package mcp

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/mcp"
	"time"
)

func (s *Server) registerCalculationTool() {
	calculatorTool := mcp.Tool{
		Name:        "calculator",
		Description: "execute calculation",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"operation": map[string]any{
					"type":        "string",
					"description": "operation (add, subtract, multiply, divide)",
					"enum":        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": map[string]any{
					"type":        "number",
					"description": "first number",
				},
				"b": map[string]any{
					"type":        "number",
					"description": "second number",
				},
			},
			Required: []string{"operation", "a", "b"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Operation string  `json:"operation"`
				A         float64 `json:"a"`
				B         float64 `json:"b"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("parse params error: %v", err)
			}

			// 执行操作
			var result float64
			switch req.Operation {
			case "add":
				result = req.A + req.B
			case "subtract":
				result = req.A - req.B
			case "multiply":
				result = req.A * req.B
			case "divide":
				if req.B == 0 {
					return nil, fmt.Errorf("divide by zero")
				}
				result = req.A / req.B
			default:
				return nil, fmt.Errorf("unknow operation: %s", req.Operation)
			}

			return map[string]any{
				"expression": fmt.Sprintf("%g %s %g", req.A, getOperationSymbol(req.Operation), req.B),
				"result":     result,
			}, nil
		},
	}

	s.mcpServer.RegisterPrompt(mcp.Prompt{
		Name:        "dynamic-prompt",
		Description: "A prompt that uses a handler to generate dynamic content",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "username",
				Description: "User's name for personalized greeting",
				Required:    true,
			},
			{
				Name:        "topic",
				Description: "Topic of expertise",
				Required:    true,
			},
		},
		Handler: func(ctx context.Context, args map[string]string) ([]mcp.PromptMessage, error) {
			var req struct {
				Username string `json:"username"`
				Topic    string `json:"topic"`
			}

			if err := mcp.ParseArguments(args, &req); err != nil {
				return nil, fmt.Errorf("failed to parse args: %w", err)
			}

			// Create a user message
			userMessage := mcp.PromptMessage{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Text: fmt.Sprintf("Hello, I'm %s and I'd like to learn about %s.", req.Username, req.Topic),
				},
			}

			// Create an assistant response with current time
			currentTime := time.Now().Format(time.RFC1123)
			assistantMessage := mcp.PromptMessage{
				Role: mcp.RoleAssistant,
				Content: mcp.TextContent{
					Text: fmt.Sprintf("Hello %s! I'm an AI assistant and I'll help you learn about %s. The current time is %s.",
						req.Username, req.Topic, currentTime),
				},
			}

			// Return both messages as a conversation
			return []mcp.PromptMessage{userMessage, assistantMessage}, nil
		},
	})

	if err := s.mcpServer.RegisterTool(calculatorTool); err != nil {
		logx.Errorf("register calculator tool error: %v", err)
	}
}

func getOperationSymbol(op string) string {
	switch op {
	case "add":
		return "+"
	case "subtract":
		return "-"
	case "multiply":
		return "×"
	case "divide":
		return "÷"
	default:
		return op
	}
}
