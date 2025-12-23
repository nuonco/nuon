package nuon

import (
	"github.com/go-openapi/runtime"
	"github.com/nuonco/nuon-go/client/operations"
)

type responseHeaderReader struct {
	resp       runtime.ClientResponse
	downstream runtime.ClientResponseReader
}

var _ runtime.ClientResponseReader = (*responseHeaderReader)(nil)

func newResponseHeaderReader(downstream runtime.ClientResponseReader) *responseHeaderReader {
	return &responseHeaderReader{
		downstream: downstream,
	}
}

func (r *responseHeaderReader) ReadResponse(cRes runtime.ClientResponse, con runtime.Consumer) (interface{}, error) {
	r.resp = cRes
	return r.downstream.ReadResponse(cRes, con)
}

func (r *responseHeaderReader) ClientOption() operations.ClientOption {
	return func(co *runtime.ClientOperation) {
		co.Reader = r
	}
}

func (r *responseHeaderReader) GetHeader(key string) string {
	if r.resp != nil {
		return r.resp.GetHeader(key)
	}
	return ""
}
