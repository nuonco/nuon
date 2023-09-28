package components

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, compID string, asJSON bool) {
	if asJSON == true {
		res, err := s.api.DeleteComponent(ctx, compID)
		if err != nil {
			fmt.Printf("failed to delete component: %s", err)
			return
		}

		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: compID, Deleted: res}
		j, _ := json.Marshal(r)
		fmt.Println(string(j))
	} else {
		view := ui.NewDeleteView("component", compID)

		_, err := s.api.DeleteComponent(ctx, compID)
		if err != nil {
			view.Fail(err)
			return
		}
		view.Success()
	}
}
