import { api } from '@/lib/api'
import type { TInstallResource } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallResources = ({
  orgId,
  installId,
  installComponentId,
  kind,
  namespace,
  health,
  provider,
}: {
  orgId: string
  installId: string
  installComponentId?: string
  kind?: string
  namespace?: string
  health?: string
  provider?: string
}) =>
  api<TInstallResource[]>({
    orgId,
    path: `installs/${installId}/resources${buildQueryParams({
      install_component_id: installComponentId,
      kind,
      namespace,
      health,
      provider,
    })}`,
  })
