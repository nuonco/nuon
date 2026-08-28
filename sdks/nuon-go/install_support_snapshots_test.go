package nuon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallSupportSnapshotRequests(t *testing.T) {
	const installID = "vinst-1"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		require.Equal(t, "org-1", r.Header.Get("X-Nuon-Org-ID"))
		switch {
		case r.Method == http.MethodPost:
			require.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, []byte("archive"), body)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id":"ags-1","install_id":"vinst-1"}`)
		case r.URL.Path == "/v1/installs/vinst-1/support-snapshots/ags-1":
			fmt.Fprint(w, `{"id":"ags-1","install_id":"vinst-1"}`)
		default:
			fmt.Fprint(w, `[{"id":"ags-1","install_id":"vinst-1"}]`)
		}
	}))
	t.Cleanup(server.Close)

	client, err := New(WithURL(server.URL), WithAuthToken("token"), WithOrgID("org-1"))
	require.NoError(t, err)

	created, err := client.CreateInstallSupportSnapshot(context.Background(), installID, bytes.NewBufferString("archive"))
	require.NoError(t, err)
	require.Equal(t, "ags-1", created.ID)

	listed, err := client.ListInstallSupportSnapshots(context.Background(), installID)
	require.NoError(t, err)
	require.Len(t, listed, 1)

	got, err := client.GetInstallSupportSnapshot(context.Background(), installID, "ags-1")
	require.NoError(t, err)
	require.Equal(t, "ags-1", got.ID)
}
