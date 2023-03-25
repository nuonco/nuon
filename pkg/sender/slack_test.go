package sender

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSend_slack(t *testing.T) {
	expectedMsg := "test slack message"
	tests := map[string]struct {
		handlerFn   func(t *testing.T) func(http.ResponseWriter, *http.Request)
		errExpected error
	}{
		"should return successfully when 200": {
			handlerFn: func(t *testing.T) func(http.ResponseWriter, *http.Request) {
				return func(rw http.ResponseWriter, r *http.Request) {
					assert.NotNil(t, r.Body)
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Contains(t, string(body), "text")

					rw.WriteHeader(http.StatusOK)
				}
			},
		},
		"should pass in a json request": {
			handlerFn: func(t *testing.T) func(http.ResponseWriter, *http.Request) {
				return func(rw http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					jsn := &struct{ Text string }{}
					err = json.Unmarshal(body, jsn)
					assert.NoError(t, err)

					rw.WriteHeader(http.StatusOK)
				}
			},
		},
		"should pass in appropriate json": {
			handlerFn: func(t *testing.T) func(http.ResponseWriter, *http.Request) {
				return func(rw http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					jsn := &struct{ Text string }{}
					err = json.Unmarshal(body, jsn)
					assert.NoError(t, err)

					m := make(map[string]interface{})
					err = json.Unmarshal(body, &m)
					assert.NoError(t, err)
					assert.Contains(t, m, "text")
					assert.Equal(t, m["text"], "test slack message")

					rw.WriteHeader(http.StatusOK)
				}
			},
		},

		"should error when server returns anything but 200": {
			handlerFn: func(t *testing.T) func(http.ResponseWriter, *http.Request) {
				return func(rw http.ResponseWriter, r *http.Request) {
					rw.WriteHeader(http.StatusBadRequest)
				}
			},
			errExpected: errors.New("unsuccessful return status"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fn := test.handlerFn(t)
			svr := httptest.NewServer(http.HandlerFunc(fn))
			defer svr.Close()
			s := &slackNotifier{
				webhookURL: svr.URL,
			}

			err := s.Send(context.Background(), expectedMsg)
			if test.errExpected != nil {
				assert.ErrorContains(t, test.errExpected, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestNewSlackSender(t *testing.T) {
	validURL := "https://hooks.slack.com/webhook/test"
	tests := map[string]struct {
		maybeURL    string
		errExpected error
	}{
		"happy path": {
			maybeURL: validURL,
		},
		"invalid url scheme": {
			maybeURL:    "http://hooks.slack.com",
			errExpected: errInvalidURL,
		},
		"not a slack url": {
			maybeURL:    "https://notslack.com",
			errExpected: errInvalidURL,
		},
		"unparseable url": {
			maybeURL:    "htt🍆ps://invalid",
			errExpected: errInvalidURL,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s, err := NewSlackSender(test.maybeURL)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, s)
			assert.NotEmpty(t, s.webhookURL)
		})
	}
}
