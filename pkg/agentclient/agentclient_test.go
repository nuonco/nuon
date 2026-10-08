package agentclient

import "testing"

func TestDetectEnv(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		want string
		ok   bool
	}{
		{
			name: "cursor agent flag",
			env:  []string{"CURSOR_AGENT=1"},
			want: Cursor,
			ok:   true,
		},
		{
			name: "cursor invoked as agent",
			env:  []string{"CURSOR_INVOKED_AS=agent"},
			want: Cursor,
			ok:   true,
		},
		{
			name: "claude code",
			env:  []string{"CLAUDECODE=1"},
			want: Claude,
			ok:   true,
		},
		{
			name: "cursor wins when both are set",
			env:  []string{"CLAUDECODE=1", "CURSOR_AGENT=1"},
			want: Cursor,
			ok:   true,
		},
		{
			name: "neither",
			env:  []string{"TERM=xterm", "CURSOR_AGENT=0", "CURSOR_INVOKED_AS=user"},
			ok:   false,
		},
		{
			name: "override cursor",
			env:  []string{"NUON_AGENT_CLIENT=cursor", "CLAUDECODE=1"},
			want: Cursor,
			ok:   true,
		},
		{
			name: "override claude",
			env:  []string{"NUON_AGENT_CLIENT=claude", "CURSOR_AGENT=1"},
			want: Claude,
			ok:   true,
		},
		{
			name: "off disables ambient detection",
			env:  []string{"NUON_AGENT_CLIENT=off", "CURSOR_AGENT=1", "CLAUDECODE=1"},
			ok:   false,
		},
		{
			name: "off is case insensitive",
			env:  []string{"NUON_AGENT_CLIENT=OFF", "CURSOR_INVOKED_AS=agent"},
			ok:   false,
		},
		{
			name: "ai agent slug",
			env:  []string{"AI_AGENT=amp"},
			want: "amp",
			ok:   true,
		},
		{
			name: "ai agent drops version",
			env:  []string{"AI_AGENT=claude-code@2.1.0"},
			want: "claude-code",
			ok:   true,
		},
		{
			name: "product variables win over ai agent",
			env:  []string{"AI_AGENT=amp", "CLAUDECODE=1"},
			want: Claude,
			ok:   true,
		},
		{
			name: "nuon override wins over ai agent",
			env:  []string{"NUON_AGENT_CLIENT=cursor", "AI_AGENT=amp"},
			want: Cursor,
			ok:   true,
		},
		{
			name: "off disables ai agent",
			env:  []string{"NUON_AGENT_CLIENT=off", "AI_AGENT=amp"},
			ok:   false,
		},
		{
			name: "non-slug ai agent falls through",
			env:  []string{"AI_AGENT=1", "CURSOR_AGENT=1"},
			want: Cursor,
			ok:   true,
		},
		{
			name: "boolean ai agent falls through",
			env:  []string{"AI_AGENT=true", "CLAUDECODE=1"},
			want: Claude,
			ok:   true,
		},
		{
			name: "empty ai agent falls through",
			env:  []string{"AI_AGENT=", "CLAUDECODE=1"},
			want: Claude,
			ok:   true,
		},
		{
			name: "amp before claude",
			env:  []string{"AMP_CURRENT_THREAD_ID=thread-1", "CLAUDECODE=1"},
			want: Amp,
			ok:   true,
		},
		{
			name: "agent amp",
			env:  []string{"AGENT=amp"},
			want: Amp,
			ok:   true,
		},
		{
			name: "cursor extension host",
			env:  []string{"CURSOR_EXTENSION_HOST_ROLE=agent-exec"},
			want: Cursor,
			ok:   true,
		},
		{
			name: "cursor trace id",
			env:  []string{"CURSOR_TRACE_ID=trace"},
			want: Cursor,
			ok:   true,
		},
		{
			name: "gemini",
			env:  []string{"GEMINI_CLI=1"},
			want: Gemini,
			ok:   true,
		},
		{
			name: "codex thread",
			env:  []string{"CODEX_THREAD_ID=thread"},
			want: Codex,
			ok:   true,
		},
		{
			name: "codex sandbox",
			env:  []string{"CODEX_SANDBOX=seatbelt"},
			want: Codex,
			ok:   true,
		},
		{
			name: "antigravity",
			env:  []string{"ANTIGRAVITY_AGENT=1"},
			want: Antigravity,
			ok:   true,
		},
		{
			name: "augment",
			env:  []string{"AUGMENT_AGENT=1"},
			want: Augment,
			ok:   true,
		},
		{
			name: "opencode",
			env:  []string{"OPENCODE_CLIENT=1"},
			want: OpenCode,
			ok:   true,
		},
		{
			name: "claude code var",
			env:  []string{"CLAUDE_CODE=1"},
			want: Claude,
			ok:   true,
		},
		{
			name: "claude child session",
			env:  []string{"CLAUDE_CODE_CHILD_SESSION=1"},
			want: Claude,
			ok:   true,
		},
		{
			name: "cowork",
			env:  []string{"CLAUDECODE=1", "CLAUDE_CODE_IS_COWORK=1"},
			want: Cowork,
			ok:   true,
		},
		{
			name: "replit",
			env:  []string{"REPL_ID=repl"},
			want: Replit,
			ok:   true,
		},
		{
			name: "copilot",
			env:  []string{"COPILOT_MODEL=gpt"},
			want: Copilot,
			ok:   true,
		},
		{
			name: "non-slug ai agent alone is not an agent",
			env:  []string{"AI_AGENT=1"},
			ok:   false,
		},
		{
			name: "boolean ai agent alone is not an agent",
			env:  []string{"AI_AGENT=true"},
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := DetectEnv(tt.env)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got.Name != tt.want {
				t.Fatalf("name = %q, want %q", got.Name, tt.want)
			}
		})
	}
}
