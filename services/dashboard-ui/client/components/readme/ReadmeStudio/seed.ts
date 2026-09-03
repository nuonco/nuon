import type { TBlock } from './types'

export const seedBlocks = (): TBlock[] => [
  {
    key: crypto.randomUUID(),
    type: 'markdown',
    content:
      '# {{.nuon.app.name}}\n\nWelcome to your **{{.nuon.app.name}}** install. This page is rendered from live install state every time you open it.',
  },
  {
    key: crypto.randomUUID(),
    type: 'status-row',
    items: [
      {
        key: crypto.randomUUID(),
        label: 'Sandbox',
        kind: 'status',
        path: 'sandbox.status',
      },
      {
        key: crypto.randomUUID(),
        label: 'Install',
        kind: 'text',
        path: 'install.name',
      },
      {
        key: crypto.randomUUID(),
        label: 'Runner',
        kind: 'status',
        path: 'runner.status',
      },
    ],
  },
  {
    key: crypto.randomUUID(),
    type: 'banner',
    theme: 'warn',
    content:
      '**Some setup is still required.** Finish the setup runbooks once the install is provisioned.',
    condition: { path: 'sandbox.status', op: 'ne', value: 'finished' },
  },
  {
    key: crypto.randomUUID(),
    type: 'markdown',
    content: '## Components',
  },
  {
    key: crypto.randomUUID(),
    type: 'table',
    sourcePath: 'components',
    limit: 5,
    emptyText: 'No components deployed yet',
    columns: [
      {
        key: crypto.randomUUID(),
        header: 'Status',
        kind: 'status',
        path: 'status',
      },
      { key: crypto.randomUUID(), header: 'Name', kind: 'text', path: 'name' },
      {
        key: crypto.randomUUID(),
        header: 'Deployed',
        kind: 'time',
        path: 'updated_at',
      },
    ],
  },
  {
    key: crypto.randomUUID(),
    type: 'markdown',
    content: '## Operations',
  },
  {
    key: crypto.randomUUID(),
    type: 'runbook',
    id: '',
    name: 'Rotate credentials',
  },
  {
    key: crypto.randomUUID(),
    type: 'component',
    id: '',
    name: 'ctl-api',
  },
  {
    key: crypto.randomUUID(),
    type: 'markdown',
    content:
      '## Access\n\nYour application is available at [{{.nuon.domain.public_domain}}](https://{{.nuon.domain.public_domain}}).',
  },
]
