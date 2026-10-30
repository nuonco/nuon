import type { TInstallDeploymentSourceDiff } from '@/types'

export const sourceDiff: TInstallDeploymentSourceDiff = {
  kind: 'component',
  name: 'acme-api',
  status: 'available',
  partial: false,
  before: {
    id: 'bld-previous',
    repository: 'acme/example',
    commit: '1111111111111111111111111111111111111111',
    root: 'services/api',
    coverage: 'complete',
  },
  after: {
    id: 'bld-selected',
    repository: 'acme/example',
    commit: '2222222222222222222222222222222222222222',
    root: 'services/api',
    coverage: 'complete',
  },
  files: [
    { path: 'settings.yaml', change: 'modified' },
    { path: 'new.yaml', change: 'added' },
    { path: 'old.yaml', change: 'removed' },
    {
      path: 'deploy.sh',
      change: 'modified',
      before_executable: false,
      after_executable: true,
    },
    { path: 'logo.png', change: 'modified' },
  ],
}

export const sourceDiffContents: Record<
  string,
  NonNullable<TInstallDeploymentSourceDiff['selected_file']>
> = {
  'settings.yaml': {
    path: 'settings.yaml',
    before: 'replicas: 2\n',
    after: 'replicas: 3\n',
  },
  'new.yaml': { path: 'new.yaml', after: 'enabled: true\n' },
  'old.yaml': { path: 'old.yaml', before: 'legacy: true\n' },
  'deploy.sh': {
    path: 'deploy.sh',
    before: '#!/bin/sh\necho deploy\n',
    after: '#!/bin/sh\necho deploy\n',
  },
  'logo.png': {
    path: 'logo.png',
    unavailable_reason:
      'This file cannot be displayed as text. It may be binary or exceed the size limit.',
  },
}
