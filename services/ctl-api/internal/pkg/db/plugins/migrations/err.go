package migrations

import "fmt"

type MigrationErr struct {
	Model string
	Name  string
	Err   error
}

func (m MigrationErr) Error() string {
	return fmt.Sprintf("error applying %s migration to model %s: %s", m.Name, m.Model, m.Err.Error())
}

func (m MigrationErr) Unwrap() error {
	return m.Err
}
