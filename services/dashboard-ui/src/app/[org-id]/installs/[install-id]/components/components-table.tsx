'use server'

import { InstallComponentsTable as Table } from '@/components/install-components/InstallComponentsTable'
import { getInstallComponents, getAppConfigById } from '@/lib'
import type { TInstall } from '@/types'

const LIMIT = 10

export async function InstallComponentsTable({
  install,
  installId,
  limit = LIMIT,
  orgId,
  offset,
  q,
  types,
}: {
  install: TInstall
  installId: string
  orgId: string
  offset: string
  limit?: number
  q?: string
  types?: string
}) {
  const [
    { data: components, error, headers },
    { data: config, error: configError },
  ] = await Promise.all([
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

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  const componentDeps = components?.map((ic) => ({
    id: ic?.id,
    component_id: ic?.component_id,
    dependencies: config?.component_config_connections?.find(
      (c) => c?.component_id === ic?.component_id
    )?.component_dependency_ids,
  }))

  if (error && !components) {
    return (
      <div>
        <p>Could not load your components.</p>
        <p>{error.error}</p>
      </div>
    )
  }

  return (
    <Table
      components={components}
      deps={componentDeps}
      pagination={pagination}
      shouldPoll
    />
  )
}
