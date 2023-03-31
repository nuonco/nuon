package services

import (
	"context"
	"fmt"
)

func (c *commands) Run(ctx context.Context, svcName string) error {
	fmt.Println("hello world nuonctl dev run")
	return nil
}
