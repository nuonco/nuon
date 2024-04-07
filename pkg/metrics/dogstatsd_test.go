package metrics

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_writer_Incr(t *testing.T) {
	key := uuid.NewString()

	tests := map[string]struct {
		client func(*testing.T, *gomock.Controller) dogstatsdClient
		opts   func() []writerOption
		tags   []string
	}{
		"happy path": {
			opts: func() []writerOption {
				return []writerOption{
					WithTags("key:value"),
				}
			},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				client.EXPECT().Incr(key, []string{"key:value"}, float64(1)).Return(nil)
				return client
			},
		},
		"happy path with tags": {
			opts: func() []writerOption {
				return []writerOption{}
			},
			tags: []string{"key:value"},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				client.EXPECT().Incr(key, []string{"key:value"}, float64(1)).Return(nil)
				return client
			},
		},
		"disabled": {
			opts: func() []writerOption {
				return []writerOption{
					WithDisable(true),
				}
			},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				return client
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			mockCtl := gomock.NewController(t)

			opts := test.opts()
			client := test.client(t, mockCtl)

			writer, err := New(v, opts...)
			assert.NoError(t, err)
			writer.client = client

			writer.Incr(key, test.tags)
		})
	}
}

func Test_writer_Decr(t *testing.T) {
	key := uuid.NewString()

	tests := map[string]struct {
		client func(*testing.T, *gomock.Controller) dogstatsdClient
		opts   func() []writerOption
		tags   []string
	}{
		"happy path": {
			opts: func() []writerOption {
				return []writerOption{
					WithTags("key:value"),
				}
			},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				client.EXPECT().Decr(key, []string{"key:value"}, float64(1)).Return(nil)
				return client
			},
		},
		"happy path with tags": {
			opts: func() []writerOption {
				return []writerOption{}
			},
			tags: []string{"key:value"},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				client.EXPECT().Decr(key, []string{"key:value"}, float64(1)).Return(nil)
				return client
			},
		},
		"disabled": {
			opts: func() []writerOption {
				return []writerOption{
					WithDisable(true),
				}
			},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				return client
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			mockCtl := gomock.NewController(t)

			opts := test.opts()
			client := test.client(t, mockCtl)

			writer, err := New(v, opts...)
			assert.NoError(t, err)
			writer.client = client

			writer.Decr(key, test.tags)
		})
	}
}

func Test_writer_Timing(t *testing.T) {
	key := uuid.NewString()
	value := time.Second

	tests := map[string]struct {
		client func(*testing.T, *gomock.Controller) dogstatsdClient
		opts   func() []writerOption
		tags   []string
	}{
		"happy path": {
			opts: func() []writerOption {
				return []writerOption{
					WithTags("key:value"),
				}
			},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				client.EXPECT().Timing(key, value, []string{"key:value"}, defaultRate).Return(nil)
				return client
			},
		},
		"happy path with tags": {
			opts: func() []writerOption {
				return []writerOption{}
			},
			tags: []string{"key:value"},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				client.EXPECT().Timing(key, value, []string{"key:value"}, defaultRate).Return(nil)
				return client
			},
		},
		"disabled": {
			opts: func() []writerOption {
				return []writerOption{
					WithDisable(true),
				}
			},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				return client
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			mockCtl := gomock.NewController(t)

			opts := test.opts()
			client := test.client(t, mockCtl)

			writer, err := New(v, opts...)
			assert.NoError(t, err)
			writer.client = client

			writer.Timing(key, value, test.tags)
		})
	}
}

func Test_writer_Event(t *testing.T) {
	ev := generics.GetFakeObj[*statsd.Event]()

	tests := map[string]struct {
		client func(*testing.T, *gomock.Controller) dogstatsdClient
		opts   func() []writerOption
	}{
		"happy path": {
			opts: func() []writerOption {
				return []writerOption{
					WithTags("key:value"),
				}
			},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				client.EXPECT().Event(ev).Return(nil)
				return client
			},
		},
		"disabled": {
			opts: func() []writerOption {
				return []writerOption{
					WithDisable(true),
				}
			},
			client: func(t *testing.T, mockCtl *gomock.Controller) dogstatsdClient {
				client := NewMockdogstatsdClient(mockCtl)
				return client
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			mockCtl := gomock.NewController(t)

			opts := test.opts()
			client := test.client(t, mockCtl)

			writer, err := New(v, opts...)
			assert.NoError(t, err)
			writer.client = client

			writer.Event(ev)
		})
	}
}
