package steps

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"go.uber.org/zap"
)

func (m Model) fetchLogsCmd() tea.Msg {
	var logStreamID string
	if m.run != nil && m.run.RunnerJob != nil {
		logStreamID = m.run.RunnerJob.LogStreamID
	}

	if logStreamID == "" {
		return logsFetchedMsg{logs: nil, logStream: nil, err: nil}
	}

	logStream, err := m.api.GetLogStream(m.ctx, logStreamID)
	if err != nil {
		return logsFetchedMsg{logs: nil, logStream: nil, err: err}
	}

	logs, err := m.api.LogStreamReadLogs(m.ctx, logStreamID, m.logsCursor, "", nil)
	return logsFetchedMsg{logs: logs, logStream: logStream, err: err}
}

func (m *Model) handleLogsFetched(msg logsFetchedMsg) {
	if msg.logs != nil {
		m.log.Info("handling logs fetched", zap.Int("logs", len(msg.logs)))
	}
	m.loadingLogs = false

	if msg.err != nil {
		m.logsFetchError = msg.err
		return
	}

	m.logsFetchError = nil

	if msg.logStream != nil {
		m.logStream = msg.logStream
	}

	for _, log := range msg.logs {
		if log == nil {
			continue
		}

		var stepName string
		if log.LogAttributes != nil {
			if name, ok := log.LogAttributes["workflow_step_name"]; ok {
				stepName = name
			}
		}

		if stepName == "" {
			continue
		}

		m.logsByStep[stepName] = append(m.logsByStep[stepName], log)
	}
	if len(m.logsByStep) > 0 {
		attrs := []zap.Field{}
		for stepName, logs := range m.logsByStep {
			attrs = append(attrs, zap.Int(stepName, len(logs)))

		}
		m.log.Info("handled logs fetched", attrs...)
	}

	m.setCursorFromLogs()

	m.setContent()
}

func (m *Model) setCursorFromLogs() {
	timestamp := "0"

	for _, logs := range m.logsByStep {
		for _, log := range logs {
			if log.Timestamp > timestamp {
				timestamp = log.Timestamp
			}
		}
	}

	if parsedTime, err := time.Parse(time.RFC3339, timestamp); err == nil {
		cursor := fmt.Sprintf("%d", parsedTime.UnixNano())
		m.logsCursor = cursor
	} else {
		m.logsCursor = "0"
	}
}
