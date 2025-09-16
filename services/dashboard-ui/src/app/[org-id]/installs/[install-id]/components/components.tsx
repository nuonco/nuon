import {
  InstallComponentsTable,
  NoComponents,
  Notice,
  Pagination,
} from '@/components'
import { api } from '@/lib/api'
import type { TInstall, TInstallComponentSummary } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const InstallComponents = async ({
  installId,
  orgId,
  limit = 10,
  offset,
  q,
  types,
}: {
  install: TInstall
  installId: string
  orgId: string
  limit?: number
  offset?: string
  q?: string
  types?: string
}) => {
  const params = buildQueryParams({ offset, limit, q, types })
  const { data, error, headers } = await api<TInstallComponentSummary[]>({
    orgId,
    path: `installs/${installId}/components/summary${params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return error ? (
    <Notice>Can&apos;t load install components: {error?.error}</Notice>
  ) : data ? (
    <div className="flex flex-col gap-4">
      <InstallComponentsTable
        initInstallComponents={data.sort((a, b) =>
          a?.component_id.localeCompare(b.component_id)
        )}
        offset={offset}
        limit={limit}
        types={types}
        q={q}
        shouldPoll
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  ) : (
    <NoComponents />
  )
}
