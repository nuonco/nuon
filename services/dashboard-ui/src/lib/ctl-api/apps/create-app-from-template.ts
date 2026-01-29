import { api } from '@/lib/api'
import type { TApp } from '@/types'

export type TCreateAppFromTemplateBody = {
  template: string
}

export const createAppFromTemplate = ({
  body,
  orgId,
}: {
  body: TCreateAppFromTemplateBody
  orgId: string
}) =>
  api<TApp>({
    body,
    method: 'POST',
    path: 'apps/from-template',
    orgId,
  })
