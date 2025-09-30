package logs

import (
	"fmt"
	"time"
)

func (m *model) getLogStream() error {
	// use the m.api and m.logStream
	logStream, err := m.api.GetLogStream(m.ctx, m.logstream_id)
	if err != nil {
		return err
	}
	m.logStream = logStream
	return nil
}

func (m *model) getLogs() error {
	m.loading = true
	m.setMessage(fmt.Sprintf("[data] fetching logs cursor:%s", m.logsCursor), "info")
	// get the next page of logs
	logs, err := m.api.LogStreamReadLogs(m.ctx, m.logstream_id, m.logsCursor)
	if err != nil {
		return err
	}

	// TODO(fd): use a dict
	for _, log := range logs {
		m.logs[log.ID] = log
	}
	m.loading = false
	return nil
}

func (m *model) setCursorFromLogs() {
	timestamp := "0"
	for _, v := range m.logs {
		if v.Timestamp > timestamp {
			timestamp = v.Timestamp
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

func (m *model) getLatestLogs() {
	// TODO(fd): we have a bit of an issue where the log stream is closed but not all the logs
	// have flushed. we should handle this.
	if m.logStream != nil && !m.logStream.Open {
		return
	}
	m.setMessage(fmt.Sprintf("[%s] getting latest data", time.Now().String()), "info")
	m.getLogs()
	m.getLogStream()
	m.setCursorFromLogs()

	rows := m.prepareRows()

	m.table.SetRows(rows)
}
