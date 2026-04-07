export default {
  title: 'LogStream/LogPanelHeading',
}

import { LogPanelHeading } from './LogPanelHeading'
import type { TOTELLog } from '@/types'

const baseLog: Partial<TOTELLog> = {
  id: 'log-1',
  timestamp: new Date(Date.now() - 120000).toISOString(),
}

export const Info = () => (
  <LogPanelHeading
    log={{ ...baseLog, severity_number: 12, severity_text: 'INFO' } as TOTELLog}
  />
)

export const Error = () => (
  <LogPanelHeading
    log={{ ...baseLog, severity_number: 20, severity_text: 'ERROR' } as TOTELLog}
  />
)

export const Warn = () => (
  <LogPanelHeading
    log={{ ...baseLog, severity_number: 16, severity_text: 'WARN' } as TOTELLog}
  />
)
