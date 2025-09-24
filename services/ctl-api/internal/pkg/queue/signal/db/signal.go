package signaldb

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

type SignalData struct {
	Signal signal.Signal
}

func (s SignalData) Value() (driver.Value, error) {
	if s.Signal == nil {
		return nil, nil
	}

	return json.Marshal(signalJSON{
		Type: s.Signal.Type(),
		Data: s.Signal,
	})
}

func (s *SignalData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type")
	}

	var out anyJSON
	if err := json.Unmarshal(bytes, &out); err != nil {
		return errors.Wrap(err, "unable to convert from wire format to any json object type")
	}

	obj, err := catalog.NewFromType(out.Type)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(out.Data, &obj); err != nil {
		return err
	}
	s.Signal = obj

	return nil
}

func (SignalData) GormDataType() string {
	return "jsonb"
}
