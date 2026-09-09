package activities

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v50/github"
	"go.uber.org/zap"
)

func TestUpsertPRCommentRecoversStaleCommentIDByMarker(t *testing.T) {
	var recoveredEdits, creates int
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/widgets/issues/comments/41", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method %s", r.Method)
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/repos/acme/widgets/issues/7/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode([]*github.IssueComment{{
				ID:   github.Int64(42),
				Body: github.String(PRCommentMarker("branch-1")),
			}})
		case http.MethodPost:
			creates++
			_ = json.NewEncoder(w).Encode(&github.IssueComment{ID: github.Int64(99)})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/repos/acme/widgets/issues/comments/42", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method %s", r.Method)
		}
		recoveredEdits++
		_ = json.NewEncoder(w).Encode(&github.IssueComment{ID: github.Int64(42)})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	client := github.NewClient(server.Client())
	client.BaseURL = baseURL
	client.UploadURL = baseURL

	staleID := int64(41)
	body := "updated"
	a := &Activities{l: zap.NewNop()}
	result, err := a.upsertPRComment(
		context.Background(),
		client,
		"acme",
		"widgets",
		&CreateOrUpdatePRCommentInput{
			PRNumber:          7,
			AppBranchID:       "branch-1",
			ExistingCommentID: &staleID,
		},
		&github.IssueComment{Body: &body},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.CommentID != 42 {
		t.Fatalf("comment ID = %d, want 42", result.CommentID)
	}
	if recoveredEdits != 1 {
		t.Fatalf("recovered edits = %d, want 1", recoveredEdits)
	}
	if creates != 0 {
		t.Fatalf("creates = %d, want 0", creates)
	}
}
