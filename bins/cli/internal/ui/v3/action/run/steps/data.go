package steps

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

// fetchLogsCmd fetches logs and log stream from the API
func (m Model) fetchLogsCmd() tea.Msg {
	// Get log stream ID from the run object
	var logStreamID string
	if m.run != nil && m.run.RunnerJob != nil {
		logStreamID = m.run.RunnerJob.LogStreamID
	}

	// Don't fetch if we don't have a log stream ID
	if logStreamID == "" {
		return logsFetchedMsg{logs: nil, logStream: nil, err: nil}
	}

	// Don't fetch if log stream is closed and the next cursor is the same
	// as the current cursor
	// if m.logStream != nil && !m.logStream.Open  {
	// 	return logsFetchedMsg{logs: nil, logStream: nil, err: nil}
	// }

	// Fetch log stream to check if it's still open
	logStream, err := m.api.GetLogStream(m.ctx, logStreamID)
	if err != nil {
		return logsFetchedMsg{logs: nil, logStream: nil, err: err}
	}

	// Fetch logs from the API
	logs, err := m.api.LogStreamReadLogs(m.ctx, logStreamID, m.logsCursor)
	return logsFetchedMsg{logs: logs, logStream: logStream, err: err}
}

// handleLogsFetched handles the logs fetched message
func (m *Model) handleLogsFetched(msg logsFetchedMsg) {
	if msg.logs != nil {
		m.log.Info("handling logs fetched", zap.Int("logs", len(msg.logs)))
	}
	m.loadingLogs = false

	if msg.err != nil {
		m.logsFetchError = msg.err
		return
	}

	// No error, clear any previous error
	m.logsFetchError = nil

	// Update log stream object
	if msg.logStream != nil {
		m.logStream = msg.logStream
	}

	// Process the logs and organize them by step name
	for _, log := range msg.logs {
		if log == nil {
			continue
		}

		// Get the step name from log attributes
		var stepName string
		if log.LogAttributes != nil {
			if name, ok := log.LogAttributes["workflow_step_name"]; ok {
				stepName = name
			}
		}

		// Skip logs without a step name
		if stepName == "" {
			continue
		}

		// Add log to the appropriate step's log list
		m.logsByStep[stepName] = append(m.logsByStep[stepName], log)
	}
	if len(m.logsByStep) > 0 {
		attrs := []zap.Field{}
		for stepName, logs := range m.logsByStep {
			attrs = append(attrs, zap.Int(stepName, len(logs)))

		}
		m.log.Info("handled logs fetched", attrs...)
	}

	// Update the cursor for the next fetch
	m.setCursorFromLogs()

	// Update the viewport content with new logs
	m.setContent()
}

// setCursorFromLogs updates the cursor based on the latest log timestamp
func (m *Model) setCursorFromLogs() {
	timestamp := "0"

	// Find the latest timestamp across all logs
	for _, logs := range m.logsByStep {
		for _, log := range logs {
			if log.Timestamp > timestamp {
				timestamp = log.Timestamp
			}
		}
	}

	// Convert timestamp string to nanosecond string
	if parsedTime, err := time.Parse(time.RFC3339, timestamp); err == nil {
		cursor := fmt.Sprintf("%d", parsedTime.UnixNano())
		m.logsCursor = cursor
	} else {
		// Fallback to original timestamp if parsing fails
		m.logsCursor = "0"
	}
}
