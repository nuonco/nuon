import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import {
  ComponentTypeFilterDropdown,
  type TComponentConfigTypeText,
} from '@/components/components/ComponentTypeFilter'
import { DriftedResources } from '@/components/install-alignment/DriftedResources'
import {
  hasOngoingDeployment,
  isDrifted,
  resourcesOfKinds,
} from '@/components/install-alignment/alignment-utils'
import { ManageAllDropdown } from '@/components/install-components/management/ManageAllDropdown'
import { ResourceComponentActions } from '@/components/install-components/ResourceComponentActions'
import { useInstall } from '@/hooks/use-install'
import { useInstallOverview } from '@/hooks/use-install-overview'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallComponents } from '@/lib'
import type { TComponentType, TInstallComponent } from '@/types'
import { InstallComponentDetailPanel } from './InstallComponentDetailPanel'
import { InstallComponentLatestDeploy } from './InstallComponentLatestDeploy'
import {
  InstallComponentsList,
  type TInstallComponentListItem,
} from './InstallComponentsList'

const COMPONENT_TYPES: TComponentType[] = [
  'helm_chart',
  'terraform_module',
  'kubernetes_manifest',
  'pulumi',
  'job',
]

const FILTER_TYPES: TComponentConfigTypeText[] = [
  'helm_chart',
  'terraform_module',
  'kubernetes_manifest',
  'pulumi',
]

const ENABLED_FILTER_OPTIONS = [
  { label: 'Enabled', value: 'enabled' },
  { label: 'Disabled', value: 'disabled' },
]

const LIMIT = 10

const enabledRank = (enabled?: boolean | null) => {
  if (enabled === false) return 2
  if (enabled === true) return 0
  return 1
}

export const InstallComponentsListContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addPanel } = useSurfaces()
  const installLink = useInstallLink()
  const [searchParams, setSearchParams] = useSearchParams()
  const showHealth = !!org?.features?.['component-health']
  const { data: overview, isPlaceholderData } = useInstallOverview()
  const alignmentById = useMemo(
    () =>
      new Map(
        (overview?.alignment?.resources ?? []).map((resource) => [
          resource.id,
          resource,
        ])
      ),
    [overview]
  )

  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined
  const enabledFilter = searchParams.get('enabled') || undefined
  const selectedTypes = (searchParams.get('types')?.split(',') ?? []).filter(
    (type): type is TComponentConfigTypeText =>
      FILTER_TYPES.includes(type as TComponentConfigTypeText)
  )
  const types = (selectedTypes.length ? selectedTypes : COMPONENT_TYPES).join(
    ','
  )
  const ids = overview?.alignment
    ? resourcesOfKinds(overview.alignment.resources, ['component'])
        .filter(
          (resource) =>
            !isDrifted(resource) &&
            !hasOngoingDeployment(resource) &&
            resource.category !== 'torn-down' &&
            (enabledFilter !== 'disabled' ||
              resource.category === 'disabled') &&
            (enabledFilter !== 'enabled' || resource.category !== 'disabled')
        )
        .map((resource) => resource.id)
        .sort()
        .join(',') || 'none'
    : undefined

  const { data: result, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'install-resource-components',
      org?.id,
      install?.id,
      offset,
      q,
      types,
      ids,
    ],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q,
        types,
        ids,
      }),
    refetchInterval: 20000,
    enabled: !!org?.id && !!install?.id,
  })

  const components = useMemo(() => {
    const toListItem = (
      installComponent: TInstallComponent
    ): TInstallComponentListItem => {
      const component = installComponent.component
      const componentId = component?.id ?? installComponent.component_id ?? ''
      const latestDeploy = installComponent.install_deploys?.[0]
      const appliedConfigId = installComponent.app_config_ref?.applied_config_id

      return {
        id: componentId || (installComponent.id ?? ''),
        name: component?.name ?? 'Component',
        type: component?.type,
        alignment: installComponent.id
          ? alignmentById.get(installComponent.id)?.category
          : undefined,
        enabled: installComponent.enabled,
        status: installComponent.status_v2?.status ?? installComponent.status,
        href: componentId
          ? installLink({
              installId: install?.id,
              appId: install?.app_id,
              suffix: `/components/${componentId}`,
            })
          : undefined,
        actions: component ? (
          <ResourceComponentActions
            onViewDetails={() =>
              addPanel(
                <InstallComponentDetailPanel
                  componentId={componentId}
                  disabled={installComponent.enabled === false}
                  name={component.name}
                />,
                `install-component-${componentId}`
              )
            }
            appliedConfigId={appliedConfigId}
            component={component}
            enabled={installComponent.enabled}
            currentDeployStatus={
              latestDeploy?.status_v2?.status ?? latestDeploy?.status
            }
          />
        ) : null,
        latestDeploy: componentId ? (
          <InstallComponentLatestDeploy
            componentId={componentId}
            showHealth={showHealth}
          />
        ) : null,
      }
    }

    const items = (result?.data ?? [])
      .filter(
        (component) => !hasOngoingDeployment(alignmentById.get(component.id))
      )
      .map(toListItem)
    const filtered =
      enabledFilter === 'enabled'
        ? items.filter((item) => item.enabled !== false)
        : enabledFilter === 'disabled'
          ? items.filter((item) => item.enabled === false)
          : items

    return [...filtered].sort(
      (a, b) => enabledRank(a.enabled) - enabledRank(b.enabled)
    )
  }, [
    result?.data,
    enabledFilter,
    org?.id,
    install?.id,
    install?.app_id,
    showHealth,
    addPanel,
    installLink,
    alignmentById,
  ])

  const tornDown: TInstallComponentListItem[] = resourcesOfKinds(
    overview?.alignment?.resources,
    ['component']
  )
    .filter(
      (resource) =>
        resource.category === 'torn-down' &&
        !hasOngoingDeployment(resource) &&
        enabledFilter !== 'disabled' &&
        (!selectedTypes.length ||
          selectedTypes.includes(
            resource.component_type as TComponentConfigTypeText
          )) &&
        (!q ||
          [resource.name, resource.id, resource.component_id].some((value) =>
            value?.toLowerCase().includes(q.toLowerCase())
          ))
    )
    .map((resource) => ({
      id: resource.id!,
      name: resource.name ?? 'Component',
      type: resource.component_type as TComponentType,
      alignment: 'torn-down',
      latestDeploy: null,
      href: resource.component_id
        ? installLink({
            installId: install?.id,
            appId: install?.app_id,
            suffix: `/components/${resource.component_id}`,
          })
        : undefined,
    }))

  return (
    <InstallComponentsList
      components={[...components, ...tornDown]}
      aligned={!!overview && !isPlaceholderData}
      drifted={(view) => (
        <DriftedResources
          kinds={['component']}
          noun="components"
          search={q}
          types={selectedTypes}
          enabled={enabledFilter}
          view={view}
        />
      )}
      loading={isLoading}
      filtered={!!q || !!selectedTypes.length || !!enabledFilter}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
      actions={<ManageAllDropdown />}
      search={
        <DebouncedSearchInput
          className="w-full md:w-fit"
          labelClassName="w-full md:w-fit"
          placeholder="Search by name or ID..."
        />
      }
      filterActions={
        <>
          <ComponentTypeFilterDropdown options={FILTER_TYPES} />
          <RadioFilterDropdown
            id="install-component-enabled-filter"
            label="Status"
            options={ENABLED_FILTER_OPTIONS}
            selected={enabledFilter}
            onChange={(next) => {
              setSearchParams((prev) => {
                const params = new URLSearchParams(prev)
                if (next) params.set('enabled', next)
                else params.delete('enabled')
                params.delete('offset')
                return params
              })
            }}
          />
        </>
      }
    />
  )
}
