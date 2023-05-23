package mappers

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

type callbackNoop func(context.Context) error

func (c callbackNoop) callback(ctx context.Context, log *log.Logger, ui terminal.UI, byts []byte) error {
	if err := c(ctx); err != nil {
		return fmt.Errorf("unable to execute callback: %w", err)
	}
	return nil
}
