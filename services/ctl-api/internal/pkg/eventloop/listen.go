package eventloop

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cockroachdb/errors"
)

// A SignalListener is a serializable representation of a workflow that wants to be notified
// about the completion of a particular signal.
//
// Field values represent the workflow to be notified, NOT the signal of interest.
type SignalListener struct {
	// WorkflowID is the id of the workflow that is waiting for a signal.
	WorkflowID string `json:"workflow_id"`
	// Namespace is the namespace of the workflow that is waiting for a signal.
	Namespace string `json:"namespace"`
	// SignalName is the name for the signal that the listening workflow is expecting. This value
	// should be dynamic and ephemeral.
	SignalName string `json:"signal_name"`
}

// SignalDoneMessage is a special one-off signal type that is sent by the event loop
// to listeners that have registered for notifications about a particular signal.
type SignalDoneMessage struct {
	Result any   `json:"result"`
	Error  error `json:"error"`
}

// NOTE(sdboyer) hand-implementing structs like this for everything containing an error isn't sustainable;
// consider this a one-off case which we use while working out a more robust solution
type portable struct {
	Result any    `json:"result"`
	Error  string `json:"error"`
	// Error  errors.EncodedError `json:"error"`
}

func (s SignalDoneMessage) MarshalJSON() ([]byte, error) {
	p := portable{
		Result: s.Result,
	}
	if s.Error != nil {
		// p.Error = errors.EncodeError(context.Background(), s.Error)
		p.Error = s.Error.Error()
	}
	return json.Marshal(p)
}

func (s *SignalDoneMessage) UnmarshalJSON(b []byte) error {
	if s == nil {
		*s = SignalDoneMessage{}
	}

	type inter struct {
		Result any    `json:"result"`
		Error  string `json:"error"`
	}
	var i inter
	err := json.Unmarshal(b, &i)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal SignalDoneMessage result")
	}
	s.Result = i.Result

	if i.Error != "" {
		s.Error = errors.New(i.Error)
	}

	return nil
}

// func (s *SignalDoneMessage) UnmarshalJSON(b []byte) error {
// 	if s == nil {
// 		*s = SignalDoneMessage{}
// 	}

// 	type inter struct {
// 		Result any             `json:"result"`
// 		Error  json.RawMessage `json:"error"`
// 	}
// 	var i inter
// 	err := json.Unmarshal(b, &i)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to unmarshal SignalDoneMessage result")
// 	}
// 	s.Result = i.Result

// 	var encoded errorspb.EncodedError
// 	if err := proto.Unmarshal(i.Error, &encoded); err != nil {
// 		return errors.Wrap(err, "failed to unmarshal SignalDoneMessage")
// 	}

// 	// NOTE(sdboyer) seems weird that we need to check this, are we holding the errors library wrong?
// 	// if p.Error.Error != nil {
// 	// 	s.Error = errors.DecodeError(context.Background(), p.Error)
// 	// }
// 	return nil
// }

// AppendListenerIDs appends the provided listeners to the provided signal. It uses
// reflection, and relies on a BaseSignal being embedded in the provided signal.
func AppendListenerIDs(sig Signal, listeners ...SignalListener) error {
	val := reflect.ValueOf(sig)

	// can only modify the struct if it's a pointer
	if val.Kind() != reflect.Ptr {
		return errors.Newf("%T is not a pointer type", sig)
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("signal must be a pointer to a struct")
	}

	baseSignalField := elem.FieldByName("BaseSignal")
	if !baseSignalField.IsValid() {
		return errors.Newf("eventloop.BaseSignal was not embedded in top level of %T", sig)
	}

	listenersField := baseSignalField.FieldByName("SignalListeners")
	if !listenersField.IsValid() {
		return errors.New("couldn't find SignalListeners field in BaseSignal")
	}
	if !listenersField.CanSet() {
		return errors.New("SignalListeners field cannot be modified")
	}
	currentIDs := listenersField.Interface().([]SignalListener)
	newIDs := append(currentIDs, listeners...)
	listenersField.Set(reflect.ValueOf(newIDs))

	return nil
}
