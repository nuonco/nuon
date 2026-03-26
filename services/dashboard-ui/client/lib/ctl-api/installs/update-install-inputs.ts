import { api } from '@/lib/api'
import type { TInstallInputs } from '@/types'

export type TUpdateInstallInputsBody = {
  inputs?: Record<string, string>
  role?: string
  deploy_dependents?: boolean
}

export const updateInstallInputs = ({
  body,
  installId,
  orgId,
}: {
  body: TUpdateInstallInputsBody
  installId: string
  orgId: string
}) =>
  api<TInstallInputs>({
    withHeaders: true,
    body,
    method: 'PATCH',
    orgId,
    path: `installs/${installId}/inputs`,
  })
