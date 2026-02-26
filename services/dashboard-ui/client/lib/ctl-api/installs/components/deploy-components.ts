import { api } from '@/lib/api'

export type TDeployComponentsBody = {
  plan_only?: boolean
}

export const deployComponents = ({
  installId,
  orgId,
  body,
}: {
  installId: string
  orgId: string
  body: TDeployComponentsBody
}) =>
  api<string>({
    path: `installs/${installId}/components/deploy-all`,
    method: 'POST',
    orgId,
    body,
  })
