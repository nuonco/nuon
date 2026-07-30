package triggers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/eventfilter"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type EventTestResult struct {
	Event EventTestMetadata `json:"event"`
	Rules []RuleTestResult  `json:"rules"`
}

type EventTestMetadata struct {
	ID          string    `json:"id"`
	TriggerID   string    `json:"trigger_id"`
	TriggerName string    `json:"trigger_name"`
	EventType   string    `json:"event_type"`
	ReceivedAt  time.Time `json:"received_at"`
}

type RuleTestResult struct {
	Name       string                `json:"name"`
	Trigger    string                `json:"trigger"`
	EventTypes []EventTypeTestResult `json:"event_types"`
	Filters    []FilterTestResult    `json:"filters"`
	Matched    bool                  `json:"matched"`
}

type EventTypeTestResult struct {
	EventType string `json:"event_type"`
	Matched   bool   `json:"matched"`
}

type FilterTestResult struct {
	From     string `json:"from"`
	Path     string `json:"path"`
	Op       string `json:"op"`
	Expected any    `json:"expected,omitempty"`
	Selected []any  `json:"selected,omitempty"`
	Matched  bool   `json:"matched"`
}

func SelectorError() error {
	return &ui.CLIUserError{Msg: "exactly one of --event or --last is required"}
}
func AppConfigError() error { return &ui.CLIUserError{Msg: "--app-config is required"} }

type PathRow struct {
	From  string `json:"from"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

const (
	eventPathMaxDepth       = 20
	eventPathMaxCount       = 1000
	eventPathTruncatedValue = "Additional payload paths omitted."
	tailRecentTerminalLimit = 1000
)

func sanitizeHumanText(value string) string {
	var sanitized strings.Builder
	for _, r := range value {
		if r <= 0x1f || (r >= 0x7f && r <= 0x9f) {
			if r <= 0xff {
				fmt.Fprintf(&sanitized, `\x%02x`, r)
			} else {
				fmt.Fprintf(&sanitized, `\u%04x`, r)
			}
			continue
		}
		sanitized.WriteRune(r)
	}
	return sanitized.String()
}

func sanitizeHumanRows(rows [][]string) [][]string {
	for row := range rows {
		for column := range rows[row] {
			rows[row][column] = sanitizeHumanText(rows[row][column])
		}
	}
	return rows
}

func (s *Service) resolveTriggerID(ctx context.Context, trigger string) (string, error) {
	if trigger == "" {
		return "", &ui.CLIUserError{Msg: "--trigger is required"}
	}
	triggers, err := s.api.ListTriggers(ctx)
	if err != nil {
		return "", err
	}
	var named []string
	for _, candidate := range triggers {
		if candidate.ID == trigger {
			return candidate.ID, nil
		}
		if candidate.Name == trigger {
			named = append(named, candidate.ID)
		}
	}
	switch len(named) {
	case 0:
		return "", &ui.CLIUserError{Msg: fmt.Sprintf("trigger %q not found", trigger)}
	case 1:
		return named[0], nil
	default:
		return "", &ui.CLIUserError{Msg: fmt.Sprintf("multiple triggers are named %q; use a trigger ID", trigger)}
	}
}

func (s *Service) List(ctx context.Context, filters models.TriggerEventListQuery, asJSON bool) error {
	triggerID, err := s.resolveTriggerID(ctx, filters.Trigger)
	if err != nil {
		return ui.PrintError(err)
	}
	filters.Trigger = triggerID
	page, err := s.api.SearchTriggerEvents(ctx, filters)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(page)
		return nil
	}
	data := [][]string{{"ID", "TRIGGER", "EVENT TYPE", "ROUTING STATUS", "MATCHES", "DISPATCHES", "RECEIVED AT"}}
	for _, event := range page.Items {
		data = append(data, []string{event.ID, triggerLabel(event.TriggerID, event.TriggerName), event.EventType, event.RoutingStatus, strconv.Itoa(event.MatchCount), strconv.Itoa(event.DispatchCount), event.ReceivedAt.String()})
	}
	ui.NewListView().Render(sanitizeHumanRows(data))
	if page.NextCursor != "" {
		ui.NewGetView().Render(sanitizeHumanRows([][]string{{"next cursor", page.NextCursor}}))
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string, asJSON bool) error {
	event, err := s.api.GetTriggerEvent(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(event)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{
		{"id", event.ID}, {"trigger", triggerLabel(event.TriggerID, event.TriggerName)}, {"external id", event.ExternalID},
		{"event type", event.EventType}, {"routing status", event.RoutingStatus}, {"routing error", event.RoutingError},
		{"matches", strconv.Itoa(event.MatchCount)}, {"dispatches", strconv.Itoa(event.DispatchCount)}, {"received at", event.ReceivedAt.String()},
	}))
	return nil
}

func (s *Service) ListDispatches(ctx context.Context, limit int, eventID, cursor string, asJSON bool) error {
	page, err := s.api.ListTriggerEventDispatchesPage(ctx, limit, eventID, cursor)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(page)
		return nil
	}
	rows := [][]string{{"ID", "EVENT ID", "APP ID", "STATUS", "ATTEMPTS", "CREATED AT"}}
	for _, dispatch := range page.Items {
		rows = append(rows, []string{dispatch.ID, dispatch.TriggerEventID, dispatch.AppID, dispatch.Status, strconv.Itoa(dispatch.Attempts), dispatch.CreatedAt.String()})
	}
	ui.NewListView().Render(sanitizeHumanRows(rows))
	if page.NextCursor != "" {
		ui.NewGetView().Render(sanitizeHumanRows([][]string{{"next cursor", page.NextCursor}}))
	}
	return nil
}

func (s *Service) Replay(ctx context.Context, id string, asJSON bool) error {
	response, err := s.api.ReplayTriggerEvent(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(response)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"event id", response.EventID}, {"replay id", response.ReplayID}}))
	return nil
}

func (s *Service) Paths(ctx context.Context, id string, last bool, trigger string, asJSON bool) error {
	event, err := s.selectEvent(ctx, id, last, trigger)
	if err != nil {
		return ui.PrintError(err)
	}
	rows, err := FlattenPaths(event)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(rows)
		return nil
	}
	data := [][]string{{"FROM", "PATH", "VALUE"}}
	for _, row := range rows {
		value, _ := json.Marshal(row.Value)
		data = append(data, []string{row.From, row.Path, string(value)})
	}
	ui.NewListView().Render(sanitizeHumanRows(data))
	return nil
}

func (s *Service) Test(ctx context.Context, id string, last bool, trigger, filename string, asJSON bool) error {
	event, err := s.selectEvent(ctx, id, last, trigger)
	if err != nil {
		return ui.PrintError(err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		return ui.PrintError(err)
	}
	parseCfg := parse.ParseConfig{Filename: filename, BackendType: config.BackendTypeLocal, Template: true, V: validator.New()}
	var cfg *config.AppConfig
	if info.IsDir() {
		parseCfg.Filename = ""
		parseCfg.Dirname = filename
		parseCfg.FileProcessor = func(_ string, obj map[string]any) map[string]any { return obj }
		cfg, err = parse.ParseDir(ctx, parseCfg)
	} else {
		cfg, err = parse.Parse(parseCfg)
	}
	if err != nil {
		return ui.PrintError(err)
	}
	if err := validate.ValidateTriggers(cfg); err != nil {
		return ui.PrintError(err)
	}
	result, err := EvaluateRules(event, cfg)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	rows := [][]string{{"MATCH", "RULE / CHECK", "SELECTED"}}
	for _, rule := range result.Rules {
		rows = append(rows, []string{mark(rule.Matched), rule.Name, ""})
		for _, eventType := range rule.EventTypes {
			rows = append(rows, []string{mark(eventType.Matched), "  event_type = " + eventType.EventType, event.EventType})
		}
		for _, filter := range rule.Filters {
			selected, _ := json.Marshal(filter.Selected)
			rows = append(rows, []string{mark(filter.Matched), fmt.Sprintf("  %s %s %s", filter.From, filter.Path, filter.Op), string(selected)})
		}
	}
	ui.NewListView().Render(sanitizeHumanRows(rows))
	return nil
}

func (s *Service) selectEvent(ctx context.Context, id string, last bool, trigger string) (*models.TriggerEvent, error) {
	if !last {
		event, err := s.api.GetTriggerEvent(ctx, id)
		if err != nil {
			return nil, err
		}
		return event, nil
	}
	triggerID, err := s.resolveTriggerID(ctx, trigger)
	if err != nil {
		return nil, err
	}
	events, err := s.api.ListTriggerEvents(ctx, 1, triggerID)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, &ui.CLIUserError{Msg: "no trigger events found"}
	}
	event, err := s.api.GetTriggerEvent(ctx, events[0].ID)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func EvaluateRules(event *models.TriggerEvent, cfg *config.AppConfig) (*EventTestResult, error) {
	if event.TriggerName == "" {
		return nil, errors.New("event trigger name is absent; cannot match local rules by trigger")
	}
	var payload any
	decoder := json.NewDecoder(strings.NewReader(string(event.Payload)))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode event payload: %w", err)
	}
	result := &EventTestResult{Event: EventTestMetadata{ID: event.ID, TriggerID: event.TriggerID, TriggerName: event.TriggerName, EventType: event.EventType, ReceivedAt: event.ReceivedAt}}
	if cfg.Triggers == nil {
		return result, nil
	}
	for _, rule := range cfg.Triggers.Rules {
		if rule.Trigger != event.TriggerName {
			continue
		}
		ruleResult := RuleTestResult{Name: rule.Name, Trigger: rule.Trigger, Matched: true}
		if len(rule.EventTypes) > 0 {
			ruleResult.Matched = false
			for _, expected := range rule.EventTypes {
				matched := expected == event.EventType
				ruleResult.EventTypes = append(ruleResult.EventTypes, EventTypeTestResult{EventType: expected, Matched: matched})
				ruleResult.Matched = ruleResult.Matched || matched
			}
		}
		for _, filter := range rule.Filters {
			compiled, err := eventfilter.Compile(eventfilter.Filter{From: eventfilter.Source(filter.From), Path: filter.Path, Op: eventfilter.Operator(filter.Op), Value: filter.Value})
			if err != nil {
				return nil, fmt.Errorf("compile rule %q filter: %w", rule.Name, err)
			}
			filterResult := compiled.Evaluate(payload, http.Header(event.Headers))
			ruleResult.Filters = append(ruleResult.Filters, FilterTestResult{From: filter.From, Path: filter.Path, Op: filter.Op, Expected: filter.Value, Selected: filterResult.Selected, Matched: filterResult.Matched})
			ruleResult.Matched = ruleResult.Matched && filterResult.Matched
		}
		result.Rules = append(result.Rules, ruleResult)
	}
	return result, nil
}

func (s *Service) Tail(ctx context.Context, trigger string, interval time.Duration, showRaw bool) error {
	if interval <= 0 {
		return ui.PrintError(&ui.CLIUserError{Msg: "--poll-interval must be greater than zero"})
	}
	triggerID, err := s.resolveTriggerID(ctx, trigger)
	if err != nil {
		return ui.PrintError(err)
	}
	trigger = triggerID
	seen := newTailSeen(tailRecentTerminalLimit)
	initialized := false
	for {
		events, err := s.eventsSinceSeen(ctx, trigger, seen, initialized)
		if err != nil {
			return ui.PrintError(err)
		}
		headings := []string{"ID", "TRIGGER", "EVENT TYPE", "DISPOSITION", "DISPATCHES", "RECEIVED AT"}
		if showRaw {
			headings = append(headings, "RAW BODY")
		}
		rows := [][]string{headings}
		for i := len(events) - 1; i >= 0; i-- {
			event := events[i]
			dispatchSummary := triggerEventDispatchSummary(event.Dispatches)
			fingerprint := event.RoutingStatus + "|" + dispatchSummary
			if previous, ok := seen.fingerprints[event.ID]; ok && previous == fingerprint {
				continue
			}
			seen.record(event, fingerprint)
			if !initialized {
				continue
			}
			row := []string{event.ID, triggerLabel(event.TriggerID, event.TriggerName), event.EventType, eventDisposition(event.RoutingStatus), dispatchSummary, event.ReceivedAt.String()}
			if showRaw {
				raw, err := s.api.GetTriggerEventRaw(ctx, event.ID)
				if err != nil {
					return ui.PrintError(err)
				}
				body, err := base64.StdEncoding.DecodeString(raw.RawBodyBase64)
				if err != nil {
					return ui.PrintError(fmt.Errorf("decode raw body for event %s: %w", event.ID, err))
				}
				row = append(row, string(body))
			}
			rows = append(rows, row)
		}
		initialized = true
		if len(rows) > 1 {
			ui.NewListView().Render(sanitizeHumanRows(rows))
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

type tailSeen struct {
	fingerprints   map[string]string
	terminal       map[string]bool
	recentTerminal []string
	terminalLimit  int
}

func newTailSeen(terminalLimit int) *tailSeen {
	return &tailSeen{fingerprints: make(map[string]string), terminal: make(map[string]bool), terminalLimit: terminalLimit}
}

func (s *tailSeen) record(event *models.TriggerEventSummary, fingerprint string) {
	s.fingerprints[event.ID] = fingerprint
	isTerminal := triggerEventTerminal(event)
	s.terminal[event.ID] = isTerminal
	for idx, id := range s.recentTerminal {
		if id == event.ID {
			s.recentTerminal = append(s.recentTerminal[:idx], s.recentTerminal[idx+1:]...)
			break
		}
	}
	if !isTerminal {
		return
	}
	s.recentTerminal = append(s.recentTerminal, event.ID)
	for len(s.recentTerminal) > s.terminalLimit {
		oldest := s.recentTerminal[0]
		s.recentTerminal = s.recentTerminal[1:]
		delete(s.fingerprints, oldest)
		delete(s.terminal, oldest)
	}
}

func triggerEventTerminal(event *models.TriggerEventSummary) bool {
	if event.RoutingStatus == "accepted" || event.RoutingStatus == "routing" {
		return false
	}
	for _, dispatch := range event.Dispatches {
		switch dispatch.Status {
		case "triggered", "dead_lettered", "cancelled":
		default:
			return false
		}
	}
	return true
}

func (s *Service) eventsSinceSeen(ctx context.Context, trigger string, seen *tailSeen, stopAtSeen bool) ([]*models.TriggerEventSummary, error) {
	var events []*models.TriggerEventSummary
	cursor := ""
	for {
		page, err := s.api.ListTriggerEventsPage(ctx, 100, trigger, cursor)
		if err != nil {
			return nil, err
		}
		for _, event := range page.Items {
			if _, ok := seen.fingerprints[event.ID]; stopAtSeen && ok {
				events = append(events, event)
				return events, nil
			}
			events = append(events, event)
		}
		if page.NextCursor == "" || !stopAtSeen {
			return events, nil
		}
		cursor = page.NextCursor
	}
}

func triggerEventDispatchSummary(dispatches []models.TriggerEventDispatchSummary) string {
	if len(dispatches) == 0 {
		return "—"
	}
	values := make([]string, 0, len(dispatches))
	for _, dispatch := range dispatches {
		value := dispatch.ID + "=" + dispatch.Status
		if dispatch.Error != "" {
			value += " (" + dispatch.Error + ")"
		}
		values = append(values, value)
	}
	sort.Strings(values)
	if len(values) == 0 {
		return "—"
	}
	return strings.Join(values, ", ")
}

func eventDisposition(status string) string {
	switch status {
	case "matched":
		return "ok"
	case "routing_failed":
		return "failed"
	default:
		return status
	}
}

func triggerLabel(id, name string) string {
	if name != "" {
		return name
	}
	return id
}

func mark(matched bool) string {
	if matched {
		return "✓"
	}
	return "✗"
}

func FlattenPaths(event *models.TriggerEvent) ([]PathRow, error) {
	var payload any
	decoder := json.NewDecoder(strings.NewReader(string(event.Payload)))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode event payload: %w", err)
	}
	rows := flattenPayload(payload)
	for name, values := range event.Headers {
		rows = append(rows, PathRow{From: "headers", Path: name, Value: values})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].From != rows[j].From {
			return rows[i].From < rows[j].From
		}
		return rows[i].Path < rows[j].Path
	})
	return rows, nil
}

func flattenPayload(value any) []PathRow {
	type pendingPath struct {
		value any
		path  string
		depth int
	}
	rows := make([]PathRow, 0)
	pending := []pendingPath{{value: value, path: "$"}}
	truncated := false
	for len(pending) > 0 && len(rows) < eventPathMaxCount {
		item := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if item.depth >= eventPathMaxDepth {
			switch item.value.(type) {
			case map[string]any, []any:
				truncated = true
				continue
			}
		}
		remaining := eventPathMaxCount - len(rows)
		switch current := item.value.(type) {
		case map[string]any:
			if len(current) == 0 {
				rows = append(rows, PathRow{From: "payload", Path: item.path, Value: current})
				continue
			}
			keys := make([]string, 0, len(current))
			for key := range current {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			if len(keys) > remaining {
				keys = keys[:remaining]
				truncated = true
			}
			for idx := len(keys) - 1; idx >= 0; idx-- {
				key := keys[idx]
				pending = append(pending, pendingPath{value: current[key], path: item.path + memberPath(key), depth: item.depth + 1})
			}
		case []any:
			if len(current) == 0 {
				rows = append(rows, PathRow{From: "payload", Path: item.path, Value: current})
				continue
			}
			count := min(len(current), remaining)
			if count < len(current) {
				truncated = true
			}
			for idx := count - 1; idx >= 0; idx-- {
				pending = append(pending, pendingPath{value: current[idx], path: fmt.Sprintf("%s[%d]", item.path, idx), depth: item.depth + 1})
			}
		default:
			rows = append(rows, PathRow{From: "payload", Path: item.path, Value: current})
		}
	}
	if len(pending) > 0 {
		truncated = true
	}
	if truncated {
		rows = append(rows, PathRow{From: "payload", Path: "…", Value: eventPathTruncatedValue})
	}
	return rows
}

func memberPath(key string) string {
	if isShorthandMember(key) {
		return "." + key
	}
	encoded, _ := json.Marshal(key)
	return "[" + string(encoded) + "]"
}

func isShorthandMember(key string) bool {
	for idx, r := range key {
		if (idx == 0 && r != '_' && !unicode.IsLetter(r)) || (idx > 0 && r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r)) {
			return false
		}
	}
	return key != ""
}
