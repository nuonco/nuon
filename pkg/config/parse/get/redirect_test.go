package get

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/publicnet"
)

func TestParse_RedirectsWithoutASource(t *testing.T) {
	var otherHits atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		otherHits.Add(1)
		_, _ = w.Write([]byte("internal"))
	}))
	defer other.Close()

	named := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/to-other-host":
			http.Redirect(w, r, other.URL+"/secret", http.StatusFound)
		case "/to-same-host", "/to-denied-path":
			http.Redirect(w, r, strings.Replace(r.URL.Path, "/to-", "/final-", 1), http.StatusFound)
		default:
			_, _ = w.Write([]byte("contents"))
		}
	}))
	defer named.Close()

	parse := func(path string, check func(field, rawURL string) error) (string, error) {
		obj := &testStruct{Obj: named.URL + path}
		err := Parse(context.Background(), obj, &Options{
			FieldTimeout: time.Second * 5,
			RootDir:      t.TempDir(),
			CheckRemote:  check,
		})
		return obj.Obj, err
	}

	tests := map[string]struct {
		path    string
		check   func(field, rawURL string) error
		want    string
		wantErr string
		wantIs  error
	}{
		"a redirect on the named host is followed": {
			path: "/to-same-host",
			want: "contents",
		},
		"a redirect to another host at a non-public address is refused": {
			path:   "/to-other-host",
			wantIs: publicnet.ErrNonPublic,
		},
		"a redirect target is checked like a named url": {
			path: "/to-denied-path",
			check: func(_, rawURL string) error {
				if strings.HasSuffix(rawURL, "/final-denied-path") {
					return errors.New("not allowed")
				}
				return nil
			},
			wantErr: "not allowed",
		},
		"an allowed redirect target is followed": {
			path:  "/to-same-host",
			check: func(_, _ string) error { return nil },
			want:  "contents",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			otherHits.Store(0)
			got, err := parse(tc.path, tc.check)
			switch {
			case tc.wantIs != nil:
				require.ErrorIs(t, err, tc.wantIs)
			case tc.wantErr != "":
				require.ErrorContains(t, err, tc.wantErr)
			default:
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
			require.Zero(t, otherHits.Load())
		})
	}
}
