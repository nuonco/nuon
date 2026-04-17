import { api } from '@/lib/api'
import type { TInstallConfig } from '@/types'

export type TUpdateInstallConfigBody = {
  approval_option?: 'approve-all' | 'prompt'
  vpc_nested_template_url?: string
  runner_nested_template_url?: string
  custom_nested_stacks?: Array<{
    name: string
    template_url: string
    index?: number
    parameters?: Record<string, string>
  }>
}

export async function updateInstallConfig({
  body,
  installConfigId,
  installId,
  orgId,
}: {
  body: TUpdateInstallConfigBody
  installConfigId: string
  installId: string
  orgId: string
}) {
  return api<TInstallConfig>({
    body,
    method: 'PATCH',
    orgId,
    path: `installs/${installId}/configs/${installConfigId}`,
  })
}
