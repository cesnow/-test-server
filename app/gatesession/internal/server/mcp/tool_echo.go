package mcp

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/mcp"
)

func (s *Server) registerEchoTool() {

	echoTool := mcp.Tool{
		Name:        "echo",
		Description: "Echoes back the message provided by the user",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"message": map[string]any{
					"type":        "string",
					"description": "The message to echo back",
				},
				"prefix": map[string]any{
					"type":        "string",
					"description": "Optional prefix to add to the echoed message",
					"default":     "Echo: ",
				},
			},
			Required: []string{"message"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Message string `json:"message"`
				Prefix  string `json:"prefix,optional"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			prefix := "Echo: "
			if len(req.Prefix) > 0 {
				prefix = req.Prefix
			}

			return prefix + req.Message, nil
		},
	}

	s.mcpServer.RegisterPrompt(mcp.Prompt{
		Name:        "hello",
		Description: "A simple hello prompt",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "name",
				Description: "The name to greet",
				Required:    false,
			},
		},
		Content: "Say hello to {{name}} and introduce yourself as an AI assistant.",
	})

	if err := s.mcpServer.RegisterTool(echoTool); err != nil {
		logx.Errorf("failed to register echo tool: %v", err)
	}

}
