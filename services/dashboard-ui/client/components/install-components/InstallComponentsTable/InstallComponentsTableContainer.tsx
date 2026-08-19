import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { SyncedFilterContainer } from '@/components/common/SyncedFilter'
import { ComponentTypeFilterDropdown } from '@/components/components/ComponentTypeFilter'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { useSyncedOnlyFilter } from '@/hooks/use-synced-only-filter'
import { getInstallComponents, getComponentLabelKeys } from '@/lib'
import { parseComponentOverrideInput } from '@/utils/install-utils'
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
  const { install, labelColors } = useInstall()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined
  const types = searchParams.get('types') || undefined
  const labels = searchParams.get('labels') || undefined
  const { syncedOnly } = useSyncedOnlyFilter()

  const { data: componentsResult, isLoading } = useQuery({
    queryKey: ['install-components', org?.id, install?.id, offset, q, types, labels],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q,
        types,
        labels,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const { data: removedResult } = useQuery({
    queryKey: ['install-components-removed', org?.id, install?.id, q, types, labels],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id,
        installId: install.id,
        limit: 100,
        offset: 0,
        q,
        types,
        labels,
        synced: false,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !syncedOnly,
  })

  const showHealth = !!org?.features?.['component-health']

  const { appConfig: configResult } = useInstallAppConfig()

  const components = componentsResult?.data ?? []
  const removedComponents =
    offset === 0 && !syncedOnly ? removedResult?.data ?? [] : []
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

  const configConnections = configResult?.component_config_connections
  const componentToggles = install?.install_config?.component_toggles

  const installValues = install?.install_inputs?.at(0)?.values
  const overriddenComponentNames = new Set<string>()
  Object.entries(installValues ?? {}).forEach(([name, value]) => {
    const parsed = parseComponentOverrideInput(name)
    if (
      parsed &&
      (parsed.kind === 'helm_values' || parsed.kind === 'tf_vars') &&
      value
    ) {
      overriddenComponentNames.add(parsed.component)
    }
  })

  const orgId = org?.id
  const appId = install?.app_id

  const removedRows = parseInstallComponentSummaryToTableData(
    removedComponents,
    [],
    org?.id ?? '',
    install?.id ?? '',
    configConnections,
    componentToggles,
    labelColors,
    overriddenComponentNames,
    true
  )
  const currentRows = parseInstallComponentSummaryToTableData(
    components,
    deps,
    org?.id ?? '',
    install?.id ?? '',
    configConnections,
    componentToggles,
    labelColors,
    overriddenComponentNames
  )

  return (
    <InstallComponentsTable
      data={[...removedRows, ...currentRows]}
      filterActions={
        <div className="flex items-center gap-3">
          <SyncedFilterContainer />
          {orgId && appId ? (
            <LabelFilterDropdown
              queryKey={['component-label-keys', orgId, appId]}
              queryFn={() => getComponentLabelKeys({ orgId, appId })}
            />
          ) : null}
          <ComponentTypeFilterDropdown />
        </div>
      }
      pagination={pagination}
      isLoading={isLoading}
      showHealth={showHealth}
    />
  )
}
