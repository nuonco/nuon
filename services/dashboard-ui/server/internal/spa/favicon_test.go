package spa

import (
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

func generate(t *testing.T, cfg internal.Config) string {
	t.Helper()
	return string(generateBYOCFavicon(&cfg, zap.NewNop()))
}

func TestGenerateBYOCFavicon_NotConfigured(t *testing.T) {
	if got := generateBYOCFavicon(&internal.Config{IsBYOC: true}, zap.NewNop()); got != nil {
		t.Fatalf("expected nil when no byoc favicon config set, got %q", got)
	}
	if got := generateBYOCFavicon(&internal.Config{BYOCIconText: "Rt"}, zap.NewNop()); got != nil {
		t.Fatalf("expected nil when not BYOC, got %q", got)
	}
	if got := generateBYOCFavicon(&internal.Config{IsBYOC: true, BYOCName: "acme"}, zap.NewNop()); got != nil {
		t.Fatalf("expected nil when only name is set (name does not drive the favicon), got %q", got)
	}
	if got := generateBYOCFavicon(&internal.Config{IsBYOC: true, BYOCIconText: "a\x01"}, zap.NewNop()); got != nil {
		t.Fatalf("expected nil for invalid icon text with no color, got %q", got)
	}
	if got := generateBYOCFavicon(&internal.Config{IsBYOC: true, BYOCIconText: "   "}, zap.NewNop()); got != nil {
		t.Fatalf("expected nil for whitespace-only icon text, got %q", got)
	}
}

func TestGenerateBYOCFavicon_IconText(t *testing.T) {
	tests := []struct {
		text string
		want string
		size string
	}{
		{"R", "R", `font-size="150"`},
		{"Rt", "Rt", `font-size="130"`},
		{"xyz", "xy", `font-size="130"`},
	}
	for _, tt := range tests {
		svg := generate(t, internal.Config{IsBYOC: true, BYOCIconText: tt.text})
		if !strings.Contains(svg, ">"+tt.want+"</text>") {
			t.Errorf("icon text %q: expected glyphs %q in %q", tt.text, tt.want, svg)
		}
		if !strings.Contains(svg, tt.size) {
			t.Errorf("icon text %q: expected %s in %q", tt.text, tt.size, svg)
		}
		hasFit := strings.Contains(svg, `textLength="170"`)
		if wantFit := len([]rune(tt.want)) == 2; hasFit != wantFit {
			t.Errorf("icon text %q: textLength fit attr presence = %v, want %v", tt.text, hasFit, wantFit)
		}
	}
}

func TestGenerateBYOCFavicon_IconTextRendersDimmedMark(t *testing.T) {
	svg := generate(t, internal.Config{IsBYOC: true, BYOCIconText: "Rt"})
	if !strings.Contains(svg, `<g opacity="0.4"><g transform=`) {
		t.Errorf("expected dimmed nuon mark behind glyphs, got %q", svg)
	}
	if !strings.Contains(svg, `<path d="M121.15`) {
		t.Errorf("expected nuon mark path, got %q", svg)
	}
	if !strings.Contains(svg, "<text") {
		t.Errorf("expected text element, got %q", svg)
	}
}

func TestGenerateBYOCFavicon_ColorOnlyRendersFullMark(t *testing.T) {
	svg := generate(t, internal.Config{IsBYOC: true, BYOCColor: "#0A3D62"})
	if !strings.Contains(svg, `<path d="M121.15`) {
		t.Errorf("expected nuon mark path, got %q", svg)
	}
	if strings.Contains(svg, "opacity") {
		t.Errorf("expected full-strength mark for color-only favicon, got %q", svg)
	}
	if strings.Contains(svg, "<text") {
		t.Errorf("expected no text element for color-only favicon, got %q", svg)
	}
	if !strings.Contains(svg, `fill="#0A3D62"`) {
		t.Errorf("expected custom background, got %q", svg)
	}
	if !strings.Contains(svg, `fill="#FFFFFF"`) {
		t.Errorf("expected white mark on dark background, got %q", svg)
	}
}

func TestGenerateBYOCFavicon_InvalidIconTextWithColorFallsBackToMark(t *testing.T) {
	svg := generate(t, internal.Config{IsBYOC: true, BYOCIconText: "a\x01", BYOCColor: "#0A3D62"})
	if !strings.Contains(svg, `<path d="M121.15`) || strings.Contains(svg, "<text") {
		t.Errorf("expected mark-only favicon when icon text is invalid, got %q", svg)
	}
}

func TestGenerateBYOCFavicon_XMLEscaping(t *testing.T) {
	svg := generate(t, internal.Config{IsBYOC: true, BYOCIconText: `<&`})
	if strings.Contains(svg, "><&<") {
		t.Fatalf("glyphs not escaped: %q", svg)
	}
	if !strings.Contains(svg, "&lt;&amp;") {
		t.Errorf("expected escaped glyphs in %q", svg)
	}
}

func TestGenerateBYOCFavicon_Background(t *testing.T) {
	svg := generate(t, internal.Config{IsBYOC: true, BYOCIconText: "Rt"})
	if !strings.Contains(svg, `fill="#662F9D"`) {
		t.Errorf("expected default purple background, got %q", svg)
	}

	svg = generate(t, internal.Config{IsBYOC: true, BYOCIconText: "Rt", BYOCColor: "#1A6B4A"})
	if !strings.Contains(svg, `fill="#1A6B4A"`) {
		t.Errorf("expected custom background, got %q", svg)
	}

	svg = generate(t, internal.Config{IsBYOC: true, BYOCIconText: "Rt", BYOCColor: "1A6B4A"})
	if !strings.Contains(svg, `fill="#1A6B4A"`) {
		t.Errorf("expected bare hex to be accepted and normalized, got %q", svg)
	}

	svg = generate(t, internal.Config{IsBYOC: true, BYOCIconText: "Rt", BYOCColor: "not-a-color"})
	if !strings.Contains(svg, `fill="#662F9D"`) {
		t.Errorf("expected fallback to purple for invalid color, got %q", svg)
	}
}

func TestBuildClientConfig_BYOCBadge(t *testing.T) {
	cc := buildClientConfig(&internal.Config{IsBYOC: true, BYOCName: "acme", BYOCColor: "#F5D90A"})
	if cc.BYOCName != "acme" || cc.BYOCColor != "#F5D90A" || cc.BYOCTextColor != "#000000" {
		t.Errorf("expected badge fields with black text on light color, got %+v", cc)
	}

	cc = buildClientConfig(&internal.Config{IsBYOC: true, BYOCName: "acme"})
	if cc.BYOCColor != "#662F9D" || cc.BYOCTextColor != "#FFFFFF" {
		t.Errorf("expected purple badge with white text when no color set, got %+v", cc)
	}

	cc = buildClientConfig(&internal.Config{IsBYOC: true, BYOCName: "acme", BYOCColor: "nope"})
	if cc.BYOCColor != "#662F9D" {
		t.Errorf("expected purple badge for invalid color, got %+v", cc)
	}

	cc = buildClientConfig(&internal.Config{IsBYOC: true, BYOCColor: "#F5D90A"})
	if cc.BYOCName != "" || cc.BYOCColor != "" || cc.BYOCTextColor != "" {
		t.Errorf("expected no badge fields without a name, got %+v", cc)
	}

	cc = buildClientConfig(&internal.Config{BYOCName: "acme"})
	if cc.BYOCName != "" {
		t.Errorf("expected no badge fields when not BYOC, got %+v", cc)
	}
}

func TestGenerateBYOCFavicon_ForegroundContrast(t *testing.T) {
	svg := generate(t, internal.Config{IsBYOC: true, BYOCIconText: "Rt", BYOCColor: "#F5D90A"})
	if !strings.Contains(svg, `fill="#000000"`) {
		t.Errorf("expected black foreground on light background, got %q", svg)
	}

	svg = generate(t, internal.Config{IsBYOC: true, BYOCIconText: "Rt", BYOCColor: "#1A1A2E"})
	if !strings.Contains(svg, `fill="#FFFFFF"`) {
		t.Errorf("expected white foreground on dark background, got %q", svg)
	}

	svg = generate(t, internal.Config{IsBYOC: true, BYOCIconText: "Rt", BYOCColor: "#fff"})
	if !strings.Contains(svg, `fill="#000000"`) {
		t.Errorf("expected black foreground on 3-digit white background, got %q", svg)
	}
}
