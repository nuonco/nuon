package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type SSEEvent struct {
	Event string `json:"-"`
	Data  any    `json:"data,omitempty"`
}

type Agent struct {
	provider Provider
	executor *ToolExecutor
	store    *ConversationStore
	cfg      *Config
	l        *zap.Logger
}

func NewAgent(provider Provider, executor *ToolExecutor, store *ConversationStore, cfg *Config, l *zap.Logger) *Agent {
	return &Agent{
		provider: provider,
		executor: executor,
		store:    store,
		cfg:      cfg,
		l:        l,
	}
}

type RunOpts struct {
	Conversation *Conversation
	Token        string
	OrgID        string
	OrgName      string
}

func (a *Agent) Run(ctx context.Context, opts RunOpts, sseOut chan<- SSEEvent) {
	defer close(sseOut)

	systemPrompt := BuildSystemPrompt(OrgContext{
		OrgID:   opts.OrgID,
		OrgName: opts.OrgName,
	})

	tools := AllTools()

	for turn := 0; turn < a.cfg.MaxTurns; turn++ {
		opts.Conversation.mu.Lock()
		messages := make([]Message, len(opts.Conversation.Messages))
		copy(messages, opts.Conversation.Messages)
		opts.Conversation.mu.Unlock()

		sendSSE(sseOut, "status", map[string]string{"status": "thinking"})

		stream, err := a.provider.StreamChat(ctx, ChatRequest{
			SystemPrompt: systemPrompt,
			Messages:     messages,
			Tools:        tools,
			MaxTokens:    4096,
		})
		if err != nil {
			a.l.Error("provider stream error", zap.Error(err))
			sendSSE(sseOut, "error", map[string]string{"message": "Failed to get response from AI provider."})
			return
		}

		var textBuf strings.Builder
		var toolCalls []ToolCall
		pendingTool := map[string]*ToolCall{}
		pendingToolInput := map[string]*strings.Builder{}

		for ev := range stream {
			switch ev.Type {
			case EventTextDelta:
				textBuf.WriteString(ev.Content)
				sendSSE(sseOut, "text", map[string]string{"content": ev.Content})

			case EventToolCallStart:
				sendSSE(sseOut, "status", map[string]string{"status": "acting"})
				tc := &ToolCall{ID: ev.ToolCallID, Name: ev.ToolName}
				pendingTool[ev.ToolCallID] = tc
				pendingToolInput[ev.ToolCallID] = &strings.Builder{}
				sendSSE(sseOut, "tool_call", map[string]any{
					"id":     ev.ToolCallID,
					"tool":   ev.ToolName,
					"status": "running",
				})

			case EventToolCallDelta:
				if buf, ok := pendingToolInput[ev.ToolCallID]; ok {
					buf.WriteString(ev.Content)
				}

			case EventToolCallEnd:
				if tc, ok := pendingTool[ev.ToolCallID]; ok {
					if buf, ok := pendingToolInput[ev.ToolCallID]; ok {
						tc.Args = buf.String()
					}
					if tc.Args == "" {
						tc.Args = "{}"
					}
					toolCalls = append(toolCalls, *tc)
				}

			case EventError:
				sendSSE(sseOut, "error", map[string]string{"message": "Provider error: " + ev.Content})
				return

			case EventDone:
			}
		}

		if len(toolCalls) == 0 {
			if textBuf.Len() > 0 {
				opts.Conversation.AppendMessage(Message{
					Role:    "assistant",
					Content: textBuf.String(),
				})
			}
			sendSSE(sseOut, "status", map[string]string{"status": "done"})
			return
		}

		assistantMsg := Message{
			Role:      "assistant",
			Content:   textBuf.String(),
			ToolCalls: toolCalls,
		}
		opts.Conversation.AppendMessage(assistantMsg)

		for _, tc := range toolCalls {
			result, err := a.executor.Execute(ctx, opts.Token, opts.OrgID, tc)
			isError := false
			if err != nil {
				a.l.Warn("tool execution failed",
					zap.String("tool", tc.Name),
					zap.String("tool_call_id", tc.ID),
					zap.String("args", tc.Args),
					zap.Error(err))
				result = err.Error()
				isError = true
			}

			truncated := truncateResult(result, 32*1024)

			opts.Conversation.AppendMessage(Message{
				Role: "user",
				ToolResult: &ToolResult{
					ToolCallID: tc.ID,
					Content:    truncated,
					IsError:    isError,
				},
			})

			status := "complete"
			if isError {
				status = "error"
			}
			sendSSE(sseOut, "tool_result", map[string]any{
				"id":     tc.ID,
				"tool":   tc.Name,
				"result": truncated,
				"status": status,
			})
		}
	}

	sendSSE(sseOut, "error", map[string]string{"message": fmt.Sprintf("Reached maximum turns (%d). Please continue in a new message.", a.cfg.MaxTurns)})
	sendSSE(sseOut, "status", map[string]string{"status": "done"})
}

func sendSSE(ch chan<- SSEEvent, event string, data any) {
	ch <- SSEEvent{Event: event, Data: data}
}

func truncateResult(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}

func FormatSSE(event string, data any) string {
	b, _ := json.Marshal(data)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(b))
}
