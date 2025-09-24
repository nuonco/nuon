import {
  AppComponentsTable,
  NoComponents,
  Notice,
  Pagination,
} from '@/components'
import { getComponents } from '@/lib'
import { api } from '@/lib/api'
import type { TBuild } from '@/types'

export const AppComponents = async ({
  appId,
  configId,
  orgId,
  limit = 10,
  offset,
  q,
  types,
}: {
  appId: string
  configId: string
  orgId: string
  limit?: number
  offset?: string
  q?: string
  types?: string
}) => {
  const {
    data: components,
    error,
    headers,
  } = await getComponents({
    appId,
    limit,
    offset,
    orgId,
    q,
    types,
  })
  const hydratedComponents =
    components &&
    !error &&
    (await Promise.all(
      components
        //.filter((c) => c?.type === 'helm_chart' || c?.type === 'terraform_module')
        .sort((a, b) => a?.id?.localeCompare(b?.id))
        .map(async (comp, _) => {
          const { data: build } = await api<TBuild>({
            orgId,
            path: `components/${comp?.id}/builds/latest`,
          })

          const deps = components.filter((c) =>
            comp.dependencies?.some((d) => d === c.id)
          )

          return {
            ...comp,
            latestBuild: build,
            deps,
          }
        })
    ))

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return error ? (
    <Notice>Can&apos;t load components: {error?.error}</Notice>
  ) : components ? (
    <div className="flex flex-col gap-4 w-full">
      <AppComponentsTable
        initComponents={hydratedComponents}
        appId={appId}
        configId={configId}
        orgId={orgId}
        limit={limit}
        offset={offset}
        q={q}
        types={types}
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
