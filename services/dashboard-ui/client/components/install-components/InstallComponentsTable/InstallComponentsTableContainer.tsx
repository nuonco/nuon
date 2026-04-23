import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { ComponentTypeFilterDropdown } from '@/components/components/ComponentTypeFilter'
import { ManageAllDropdown } from '@/components/install-components/management/ManageAllDropdown'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents, getAppConfig, getComponentLabelKeys } from '@/lib'
import { InstallComponentsTable, parseInstallComponentSummaryToTableData } from './InstallComponentsTable'

const LIMIT = 10

export const InstallComponentsTableContainer = ({
  pollInterval = 20000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: componentsResult, isLoading } = useQuery({
    queryKey: [
      'install-components',
      org?.id,
      install?.id,
      offset,
      searchParams.get('q'),
      searchParams.get('types'),
      searchParams.get('labels'),
    ],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q: searchParams.get('q') || undefined,
        types: searchParams.get('types') || undefined,
        labels: searchParams.get('labels') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const { data: configResult } = useQuery({
    queryKey: [
      'app-config',
      org?.id,
      install?.app_id,
      install?.app_config_id,
      'recurse',
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_config_id,
  })

  const components = componentsResult?.data ?? []
  const pagination = {
    hasNext: componentsResult?.pagination?.hasNext ?? false,
    offset,
    limit: LIMIT,
  }

  const deps = components.map((ic) => ({
    id: ic?.id,
    component_id: ic?.component_id,
    dependencies: configResult?.component_config_connections?.find(
      (c) => c?.component_id === ic?.component_id
    )?.component_dependency_ids,
  }))

  return (
    <InstallComponentsTable
      data={parseInstallComponentSummaryToTableData(
        components,
        deps,
        org?.id ?? '',
        install?.id ?? ''
      )}
      filterActions={
        <div className="flex items-center gap-3">
          <LabelFilterDropdown
            queryKey={['component-label-keys', org.id, install?.app_id]}
            queryFn={() => getComponentLabelKeys({ orgId: org.id, appId: install.app_id })}
          />
          <ComponentTypeFilterDropdown />
          <ManageAllDropdown />
        </div>
      }
      pagination={pagination}
      isLoading={isLoading}
    />
  )
}
