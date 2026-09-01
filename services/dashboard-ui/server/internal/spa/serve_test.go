package spa

import (
	"strings"
	"testing"
)

const shellHTML = `<head>
    <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
    <link rel="stylesheet" href="/assets/styles.css" data-shell="default" />
    <link rel="stylesheet" href="/assets/app.css" data-shell="default" />
    <link rel="stylesheet" href="/assets/lite.css" data-shell="lite" />
  </head>`

func TestSelectShellLinksDefault(t *testing.T) {
	got := string(selectShellLinks([]byte(shellHTML), shellDefault))

	for _, want := range []string{"/assets/styles.css", "/assets/app.css", "/favicon.svg"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q to be kept, got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "/assets/lite.css") {
		t.Errorf("expected lite stylesheet to be stripped, got:\n%s", got)
	}
}

func TestSelectShellLinksLite(t *testing.T) {
	got := string(selectShellLinks([]byte(shellHTML), shellLite))

	for _, want := range []string{"/assets/lite.css", "/favicon.svg"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q to be kept, got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"/assets/styles.css", "/assets/app.css"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("expected %q to be stripped, got:\n%s", unwanted, got)
		}
	}
}

func TestSelectShellLinksHashedFilenames(t *testing.T) {
	hashed := `<link rel="stylesheet" href="/assets/lite-a1b2c3d4.css" data-shell="lite" />
<link rel="stylesheet" href="/assets/styles-e5f6a7b8.css" data-shell="default" />`

	got := string(selectShellLinks([]byte(hashed), shellLite))

	if !strings.Contains(got, "/assets/lite-a1b2c3d4.css") {
		t.Errorf("expected hashed lite stylesheet to be kept, got:\n%s", got)
	}
	if strings.Contains(got, "/assets/styles-e5f6a7b8.css") {
		t.Errorf("expected hashed default stylesheet to be stripped, got:\n%s", got)
	}
}
