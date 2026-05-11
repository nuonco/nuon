package signaldb

import (
	"database/sql/driver"
	"encoding/json"
	stderrors "errors"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	if err := s.unmarshalSignalJSON(bytes); err != nil {
		// this was done to make this part of code rollback compatible, if a signal in db is not found in catalog
		// it should handle the case since its very possible in case of new signal additions. Here, if its a
		// signal not registered error, we return a nil signal and let caller handle the case.
		if stderrors.Is(err, catalog.ErrSignalTypeNotRegistered) {
			zap.L().Warn("signal type from DB not registered in catalog; leaving Signal nil",
				zap.Error(err))
			return nil
		}
		return err
	}
	return nil
}

// MarshalJSON implements json.Marshaler so that standard JSON serialization
// (used by temporaljson and Temporal data converters) includes the type discriminator.
func (s SignalData) MarshalJSON() ([]byte, error) {
	if s.Signal == nil {
		return []byte("null"), nil
	}

	return json.Marshal(signalJSON{
		Type: s.Signal.Type(),
		Data: s.Signal,
	})
}

// UnmarshalJSON implements json.Unmarshaler so that standard JSON deserialization
// (used by temporaljson and Temporal data converters) can reconstruct the typed signal
// from the {type, data} wire format using the catalog.
func (s *SignalData) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	return s.unmarshalSignalJSON(data)
}

func (s *SignalData) unmarshalSignalJSON(data []byte) error {
	var out anyJSON
	if err := json.Unmarshal(data, &out); err != nil {
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
