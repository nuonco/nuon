package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/pkg/errors"
	commonpb "go.temporal.io/api/common/v1"
)

// @ID						TemporalCodecDecode
// @Summary				 Decode endpoint for Temporal payloads
// @Description			Handles decoding temporal using gzip compression
// @Param					req	body commonpb.Payloads	true "Payloads to decode"
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				200	{object}	commonpb.Payloads
// @Failure				400	{object}	map[string]string
// @Failure				500	{object}	map[string]string
// @Router					/v1/general/temporal-codec/decode [post]
func (s *service) TemporalCodecDecode(ctx *gin.Context) {
	var payloadspb commonpb.Payloads
	if err := jsonpb.Unmarshal(ctx.Request.Body, &payloadspb); err != nil {
		ctx.Error(errors.Wrap(err, "unable to parse request body"))
		return
	}

	// Apply all codecs in sequence
	for _, codec := range s.codecs {
		var err error
		out, err := codec.Decode(payloadspb.Payloads)
		if err != nil {
			ctx.Error(errors.Wrap(err, "unable to decode"))
			return
		}

		payloadspb.Payloads = out
	}

	ctx.JSON(http.StatusOK, payloadspb)
}
