import { api } from '@/lib/api'

export type TTeardownComponentsBody = {
  error_behavior?: 'continue' | 'abort'
  plan_only?: boolean
}

export const teardownComponents = ({
  installId,
  orgId,
  body,
}: {
  installId: string
  orgId: string
  body: TTeardownComponentsBody
}) =>
  api<string>({
    path: `installs/${installId}/components/deploy-all`,
    method: 'POST',
    orgId,
    body,
  })
