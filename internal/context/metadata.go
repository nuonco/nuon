package context

import (
	"context"
	"strings"

	"github.com/gogo/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const userIDHeaderKey string = "x-nuon-user-id"

func ParseMetadata(ctx context.Context) (context.Context, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}

	ids, ok := meta[userIDHeaderKey]
	if !ok || len(ids) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "missing '%s' header", userIDHeaderKey)
	}
	if strings.Trim(ids[0], " ") == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty '%s' header", userIDHeaderKey)
	}

	return WithUserID(ctx, ids[0]), nil
}
