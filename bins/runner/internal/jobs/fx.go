package jobs

import (
	"fmt"

	"go.uber.org/fx"
)

func AsJobHandler(group string, f any) any {
	return fx.Annotate(
		f,
		fx.As(new(JobHandler)),
		fx.ResultTags(fmt.Sprintf(`group:"%s"`, group)),
	)
}
