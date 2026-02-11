package common

import "github.com/nuonco/nuon/sdks/nuon-go/models"

func CalculateStepProgress(steps []*models.AppWorkflowStep) (finished, pending int, pct float64) {
	if len(steps) == 0 {
		return 0, 0, 0
	}
	total := len(steps)
	finished = 0
	for _, step := range steps {
		if step.Finished {
			finished++
		}
	}
	pending = total - finished
	pct = float64(finished) / float64(total)
	return finished, pending, pct
}
