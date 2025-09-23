import { api } from '@/lib/api'
import type { TInstallInputs } from '@/types'

export type TUpdateInstallInputsBody = {
  inputs?: Record<string, string>
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
    body,
    method: 'PATCH',
    orgId,
    path: `installs/${installId}/inputs`,
  })
