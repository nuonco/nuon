import {
  InstallComponentsTable,
  NoComponents,
  Notice,
  Pagination,
} from '@/components'
import { getInstallComponents, getAppConfigById } from '@/lib'
import type { TInstall } from '@/types'

export const InstallComponents = async ({
  install,
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
  const [{ data, error, headers }, { data: config, error: configError }] =
    await Promise.all([
      getInstallComponents({
        installId,
        limit,
        offset,
        orgId,
        q,
        types,
      }),
      getAppConfigById({
        appConfigId: install?.app_config_id,
        appId: install.app_id,
        orgId,
        recurse: true,
      }),
    ])

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  const componentDeps = data?.map((ic) => ({
    id: ic?.id,
    component_id: ic?.component_id,
    dependencies: config?.component_config_connections?.find(
      (c) => c?.component_id === ic?.component_id
    )?.component_dependency_ids,
  }))

  return error || configError ? (
    <Notice>Can&apos;t load install components: {error?.error}</Notice>
  ) : data ? (
    <div className="flex flex-col gap-4">
      <InstallComponentsTable
        componentDeps={componentDeps}
        initInstallComponents={data}
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
