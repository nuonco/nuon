import { api } from '@/lib/api'
import type { TEnvAccentColor, TOrg } from '@/types'

// PUT /v1/orgs/current/env-accent-config
//
// Backend: services/ctl-api/internal/app/orgs/service/update_env_accent_config.go
// PUT-style replacement of the entire mapping: send {label_key: "", values: {}}
// to clear and render every install neutrally.
export type TUpdateEnvAccentConfigBody = {
  label_key: string
  values: Record<string, TEnvAccentColor>
}

export const updateEnvAccentConfig = ({
  body,
  orgId,
}: {
  body: TUpdateEnvAccentConfigBody
  orgId: string
}) =>
  api<TOrg>({
    body,
    method: 'PUT',
    orgId,
    path: `orgs/current/env-accent-config`,
  })
