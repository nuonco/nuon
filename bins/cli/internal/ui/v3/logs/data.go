package logs

import (
	"fmt"
	"strconv"
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
	for {
		logs, next, err := m.api.LogStreamReadLogsWithNextOffset(m.ctx, m.logstream_id, m.logsCursor, "", nil)
		if err != nil {
			m.loading = false
			return err
		}
		for _, log := range logs {
			m.logs[log.ID] = log
		}
		if next == "" || next == m.logsCursor {
			// A short page has no next header. Remember the last record so a
			// later refresh does not reprint this page.
			if len(logs) > 0 {
				if cursor := unixNanoCursor(logs[len(logs)-1].Timestamp); cursor != "" {
					m.logsCursor = cursor
				}
			}
			break
		}
		m.logsCursor = next
	}
	m.loading = false
	return nil
}

func unixNanoCursor(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		parsed, err := time.Parse(layout, timestamp)
		if err == nil {
			return strconv.FormatInt(parsed.UnixNano(), 10)
		}
	}
	return ""
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

	rows := m.prepareRows()

	m.table.SetRows(rows)
}
