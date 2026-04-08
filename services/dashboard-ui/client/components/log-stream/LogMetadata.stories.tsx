export default {
  title: 'LogStream/LogMetadata',
}

import { LogMetadata } from './LogMetadata'
import type { TOTELLog } from '@/types'

const mockLog: TOTELLog = {
  id: 'log-abc123',
  timestamp: '2024-01-15T10:30:00Z',
  severity_number: 12,
  severity_text: 'INFO',
  body: 'Deploy completed successfully for install install-xyz',
  service_name: 'runner',
  scope_name: 'oteljob',
  log_attributes: { 'install.id': 'install-xyz', 'deploy.id': 'deploy-123' },
  resource_attributes: { 'service.version': '2.1.0' },
  scope_attributes: {},
} as TOTELLog

export const Default = () => <LogMetadata log={mockLog} />
