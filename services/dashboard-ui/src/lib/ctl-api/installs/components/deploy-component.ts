import { api } from '@/lib/api'
import type { TDeploy } from '@/types'

export type TDeployComponentBody = {
  build_id: string
  deploy_dependents?: boolean
  plan_only?: boolean
}

export const deployComponent = ({
  installId,
  orgId,
  body,
}: {
  installId: string
  orgId: string
  body: TDeployComponentBody
}) =>
  api<TDeploy>({
    path: `installs/${installId}/deploys`,
    method: 'POST',
    orgId,
    body,
  })
