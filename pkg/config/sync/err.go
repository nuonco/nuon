package sync

import "fmt"

type SyncInternalErr struct {
	Description string
	Err         error
}

func (s SyncInternalErr) Error() string {
	msg := fmt.Sprintf("error syncing - %s", s.Description)
	if s.Err != nil {
		msg = fmt.Sprintf("%s %s", msg, s.Err.Error())
	}

	return msg
}

type SyncErr struct {
	Resource    string
	Description string
}

func (s SyncErr) Error() string {
	return fmt.Sprintf("unable to sync %s - %s", s.Resource, s.Description)
}

type SyncAPIErr struct {
	Resource string
	Err      error
}

func (s SyncAPIErr) Error() string {
	return fmt.Sprintf("unable to sync %s - %s", s.Resource, s.Err.Error())
}
