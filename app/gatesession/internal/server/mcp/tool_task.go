package mcp

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/mcp"
)

func (s *Server) registerTaskTool() {
	tool := mcp.Tool{
		Name:        "task",
		Description: "list tasks by project",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"project": map[string]any{
					"type":        "string",
					"description": "project name (7S, 9S)",
				},
			},
			Required: []string{"project"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Project string `json:"project"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("parse params error: %v", err)
			}

			tasksByProject := map[string][]string{
				"7S": {
					"[7S][Backend] 支援多語系",
					"[7S][Frontend] 修正報表匯出錯誤",
				},
				"9S": {
					"[9S][Frontend] 增加浮水印在圖片上",
				},
			}

			return map[string]any{
				"expression": fmt.Sprintf("Project: %s", req.Project),
				"result":     tasksByProject[req.Project],
			}, nil
		},
	}

	if err := s.mcpServer.RegisterTool(tool); err != nil {
		logx.Errorf("register task tool error: %v", err)
	}
}
