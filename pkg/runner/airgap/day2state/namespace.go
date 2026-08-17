package day2state

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const (
	RunnerNamespace  = "runner/"
	ControlNamespace = "control/"
)

type prefixedState struct {
	state  State
	prefix string
}

func WithPrefix(state State, prefix string) State {
	prefix = strings.Trim(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return &prefixedState{state: state, prefix: prefix}
}

func (s *prefixedState) key(key string) string { return s.prefix + strings.TrimPrefix(key, "/") }

func (s *prefixedState) Get(ctx context.Context, key string) ([]byte, bool, error) {
	return s.state.Get(ctx, s.key(key))
}

func (s *prefixedState) Put(ctx context.Context, key string, raw []byte) error {
	return s.state.Put(ctx, s.key(key), raw)
}

func (s *prefixedState) PutIfAbsent(ctx context.Context, key string, raw []byte) error {
	return s.state.PutIfAbsent(ctx, s.key(key), raw)
}

func (s *prefixedState) List(ctx context.Context, prefix string) ([]string, error) {
	keys, err := s.state.List(ctx, s.key(prefix))
	if err != nil {
		return nil, err
	}
	for i := range keys {
		keys[i] = strings.TrimPrefix(keys[i], s.prefix)
	}
	return keys, nil
}

type readOverlay struct {
	states []State
}

func ReadOverlay(states ...State) State { return &readOverlay{states: states} }

func (s *readOverlay) Get(ctx context.Context, key string) ([]byte, bool, error) {
	for _, state := range s.states {
		raw, found, err := state.Get(ctx, key)
		if err != nil || found {
			return raw, found, err
		}
	}
	return nil, false, nil
}

func (s *readOverlay) List(ctx context.Context, prefix string) ([]string, error) {
	seen := map[string]bool{}
	for _, state := range s.states {
		keys, err := state.List(ctx, prefix)
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			seen[key] = true
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys, nil
}

func (s *readOverlay) Put(context.Context, string, []byte) error {
	return fmt.Errorf("read overlay does not own state")
}

func (s *readOverlay) PutIfAbsent(context.Context, string, []byte) error {
	return fmt.Errorf("read overlay does not own state")
}

type legacyState struct{ state State }

func Legacy(state State) State { return &legacyState{state: state} }

func (s *legacyState) Get(ctx context.Context, key string) ([]byte, bool, error) {
	if namespacedKey(key) {
		return nil, false, nil
	}
	return s.state.Get(ctx, key)
}

func (s *legacyState) List(ctx context.Context, prefix string) ([]string, error) {
	keys, err := s.state.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
	legacy := keys[:0]
	for _, key := range keys {
		if !namespacedKey(key) {
			legacy = append(legacy, key)
		}
	}
	return legacy, nil
}

func (s *legacyState) Put(ctx context.Context, key string, raw []byte) error {
	return s.state.Put(ctx, key, raw)
}

func (s *legacyState) PutIfAbsent(ctx context.Context, key string, raw []byte) error {
	return s.state.PutIfAbsent(ctx, key, raw)
}

func namespacedKey(key string) bool {
	key = strings.TrimPrefix(key, "/")
	return strings.HasPrefix(key, RunnerNamespace) || strings.HasPrefix(key, ControlNamespace)
}
