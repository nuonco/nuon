// Mirror of the Go SubOps map in services/ctl-api/internal/pkg/interests/types.go.
// Empty `ops: []` on a ResourceCfg means "all sub-ops" — we render this list to
// let the user narrow the per-resource filter, never to enumerate the wire
// payload. Keep this in sync with the Go map; the backend validator rejects
// unknown ops outright.

import type { ResourceKind } from './types'

export const SUB_OPS: Record<ResourceKind, string[]> = {
  installs: ['provision', 'deprovision', 'reprovision'],
  components: ['deploy', 'teardown', 'drift'],
  sandboxes: ['provision', 'reprovision', 'deprovision', 'drift'],
  install_configurations: ['inputs', 'secrets'],
  runners: ['provision', 'reprovision'],
  actions: ['run'],
}

const SUB_OP_LABELS: Record<string, string> = {
  // shared lifecycle ops
  provision: 'Provision',
  deprovision: 'Deprovision',
  reprovision: 'Reprovision',
  drift: 'Drift',
  // components
  deploy: 'Deploy',
  teardown: 'Teardown',
  // install_configurations
  inputs: 'Input updates',
  secrets: 'Secret syncs',
  // actions
  run: 'Run',
}

export const labelForSubOp = (op: string): string =>
  SUB_OP_LABELS[op] ??
  op.replace(/_/g, ' ').replace(/^./, (c) => c.toUpperCase())
