export default {
  title: 'LogStream/LogPanel',
}

import { PanelStory } from '@/components/__stories__/helpers'
import { LogPanel } from './LogPanel'
import type { TOTELLog } from '@/types'

const mockLog: TOTELLog = {
  id: 'log-abc123',
  timestamp: new Date(Date.now() - 120000).toISOString(),
  severity_number: 12,
  severity_text: 'INFO',
  body: 'Deploy completed successfully for install install-xyz. All 4 components are now active.',
  service_name: 'runner',
  scope_name: 'oteljob',
  log_attributes: {
    'install.id': 'install-xyz',
    'deploy.id': 'deploy-123',
    'component.name': 'api',
  },
  resource_attributes: {
    'service.version': '2.1.0',
    'host.name': 'worker-node-01',
  },
  scope_attributes: {
    'library.name': 'nuon-runner',
    'library.version': '1.0.0',
  },
} as TOTELLog

const mockErrorLog: TOTELLog = {
  ...mockLog,
  id: 'log-error123',
  severity_number: 20,
  severity_text: 'ERROR',
  body: 'Failed to connect to database: connection timeout after 30s',
} as TOTELLog

export const InfoPanel = () => (
  <PanelStory label="Open info log panel">
    <LogPanel log={mockLog} />
  </PanelStory>
)

export const ErrorPanel = () => (
  <PanelStory label="Open error log panel">
    <LogPanel log={mockErrorLog} />
  </PanelStory>
)
