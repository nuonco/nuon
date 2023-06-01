// build.go
package models

import (
	"time"

	buildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	"google.golang.org/genproto/googleapis/type/datetime"
)

type Build struct {
	Model

	ComponentID string
	Component   Component
	CreatedByID string
	GitRef      string `json:"git_ref"`
}

// QueryBuilds and GetBuild are using it to convert the Build model to proto format
func (b Build) ToProto() *buildv1.Build {
	return &buildv1.Build{
		Id:          b.ID,
		GitRef:      b.GitRef,
		ComponentId: b.ComponentID,
		CreatedById: b.CreatedByID,
		UpdatedAt:   TimeToDatetime(b.UpdatedAt),
		CreatedAt:   TimeToDatetime(b.CreatedAt),
	}
}

// TODO: if we really need this, we can least put it in a shared lib.
func TimeToDatetime(ts time.Time) *datetime.DateTime {
	return &datetime.DateTime{
		Year:       int32(ts.Year()),
		Month:      int32(ts.Month()),
		Day:        int32(ts.Day()),
		Hours:      int32(ts.Hour()),
		Minutes:    int32(ts.Minute()),
		Seconds:    int32(ts.Second()),
		Nanos:      int32(ts.Nanosecond()),
		TimeOffset: &datetime.DateTime_UtcOffset{},
	}
}
