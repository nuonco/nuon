import { api } from '@/lib/api'
import type { TBuild } from '@/types'

type TBuildComponentRequest = {
  git_ref?: string
  use_latest?: boolean
}

export const buildComponent = ({
  componentId,
  orgId,
  body = { use_latest: true },
}: {
  componentId: string
  orgId: string
  body?: TBuildComponentRequest
}) =>
  api<TBuild>({
    path: `components/${componentId}/builds`,
    method: 'POST',
    orgId,
    body,
  })
