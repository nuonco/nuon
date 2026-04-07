export default {
  title: 'LogStream/AttributesTabs',
}

import { AttributesTabs } from './AttributesTabs'
import type { TOTELLog } from '@/types'

const mockLog: TOTELLog = {
  id: 'log-1',
  timestamp: new Date().toISOString(),
  severity_number: 12,
  severity_text: 'INFO',
  body: 'Service started successfully',
  service_name: 'runner',
  scope_name: 'oteljob',
  log_attributes: { 'http.method': 'GET', 'http.url': '/health' },
  resource_attributes: { 'service.version': '1.0.0', 'host.name': 'worker-01' },
  scope_attributes: { 'library.name': 'nuon-runner' },
} as TOTELLog

export const Default = () => <AttributesTabs log={mockLog} />

export const Empty = () => (
  <AttributesTabs
    log={{ ...mockLog, log_attributes: {}, resource_attributes: {}, scope_attributes: {} }}
  />
)
