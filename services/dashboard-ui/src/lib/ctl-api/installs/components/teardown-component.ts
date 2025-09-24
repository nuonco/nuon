import { api } from '@/lib/api'

export type TTeardownComponentBody = {
  error_behavior?: 'continue' | 'abort'
  plan_only?: boolean
}

export const teardownComponent = ({
  componentId,
  installId,
  orgId,
  body,
}: {
  componentId: string
  installId: string
  orgId: string
  body: TTeardownComponentBody
}) =>
  api<string>({
    path: `installs/${installId}/components/${componentId}/teardown`,
    method: 'POST',
    orgId,
    body,
  })
