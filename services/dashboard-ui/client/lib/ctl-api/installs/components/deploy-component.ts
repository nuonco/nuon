import { api } from '@/lib/api'
import type { TDeploy } from '@/types'

export type TDeployComponentBody = {
  build_id: string
  deploy_dependents?: boolean
  deploy_dependencies?: boolean
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
    withHeaders: true,
    path: `installs/${installId}/deploys`,
    method: 'POST',
    orgId,
    body,
  })
