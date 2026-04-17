import { api } from '@/lib/api'
import type { TInstallConfig } from '@/types'

export type TCreateInstallConfigBody = {
  approval_option: 'approve-all' | 'prompt'
  vpc_nested_template_url?: string
  runner_nested_template_url?: string
  custom_nested_stacks?: Array<{
    name: string
    template_url: string
    index?: number
    parameters?: Record<string, string>
  }>
}

export async function createInstallConfig({
  body,
  installId,
  orgId,
}: {
  body: TCreateInstallConfigBody
  installId: string
  orgId: string
}) {
  return api<TInstallConfig>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/configs`,
  })
}
