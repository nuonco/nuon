package installs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, id string, asJSON bool) {
	if asJSON == true {
		res, err := s.api.DeleteInstall(ctx, id)
		if err != nil {
			fmt.Printf("failed to delete install: %s", err)
			return
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: id, Deleted: res}
		j, _ := json.Marshal(r)
		fmt.Println(string(j))
	} else {
		view := ui.NewDeleteView("install", id)

		view.Start()
		_, err := s.api.DeleteInstall(ctx, id)
		if err != nil {
			view.Fail(err)
			return
		}
		view.Success()
	}
}
