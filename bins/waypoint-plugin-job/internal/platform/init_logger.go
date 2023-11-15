package platform

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

func (p *Platform) initLogger(ui terminal.UI) error {
	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return fmt.Errorf("unable to get output writers")
	}
	p.logger = hclog.New(&hclog.LoggerOptions{
		Name:   "job",
		Output: stdout,
	})
	return nil
}
