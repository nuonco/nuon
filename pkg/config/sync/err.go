package sync

import (
	"fmt"
)

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

func (s SyncInternalErr) Unwrap() error { return s.Err }

type SyncErr struct {
	Resource    string
	Description string
	Err         error
}

func (s SyncErr) Error() string {
	return fmt.Sprintf("unable to sync %s - %s", s.Resource, s.Description)
}

func (s SyncErr) Unwrap() error { return s.Err }

type SyncAPIErr struct {
	Resource string
	Err      error
}

func (s SyncAPIErr) Error() string {
	return fmt.Sprintf("unable to sync %s - %s", s.Resource, s.Err.Error())
}

func (s SyncAPIErr) Unwrap() error { return s.Err }
