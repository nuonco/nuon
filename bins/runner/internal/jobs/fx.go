package jobs

import (
	"fmt"

	"go.uber.org/fx"
)

func AsJobHandler(group string, f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(fmt.Sprintf(`group:"%s"`, group)),
		fx.As(new(JobHandler)),
	)
}
