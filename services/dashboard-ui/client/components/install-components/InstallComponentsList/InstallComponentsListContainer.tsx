import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import {
  ComponentTypeFilterDropdown,
  type TComponentConfigTypeText,
} from '@/components/components/ComponentTypeFilter'
import { ManageAllDropdown } from '@/components/install-components/management/ManageAllDropdown'
import { ResourceComponentActions } from '@/components/install-components/ResourceComponentActions'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents } from '@/lib'
import type { TComponentType, TInstallComponent } from '@/types'
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

const LIMIT = 10

export const InstallComponentsListContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()
  const showHealth = !!org?.features?.['component-health']

  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined
  const selectedTypes = (searchParams.get('types')?.split(',') ?? []).filter(
    (type): type is TComponentConfigTypeText =>
      FILTER_TYPES.includes(type as TComponentConfigTypeText)
  )
  const types = (selectedTypes.length ? selectedTypes : COMPONENT_TYPES).join(
    ','
  )

  const { data: result, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'install-resource-components',
      org?.id,
      install?.id,
      offset,
      q,
      types,
    ],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q,
        types,
      }),
    refetchInterval: 20000,
    enabled: !!org?.id && !!install?.id,
  })

  const toListItem = (
    installComponent: TInstallComponent
  ): TInstallComponentListItem => {
    const component = installComponent.component
    const componentId = component?.id ?? installComponent.component_id ?? ''
    const latestDeploy = installComponent.install_deploys?.[0]

    return {
      id: componentId || (installComponent.id ?? ''),
      name: component?.name ?? 'Component',
      type: component?.type,
      enabled: installComponent.enabled,
      status: installComponent.status_v2?.status ?? installComponent.status,
      actions: component ? (
        <ResourceComponentActions
          component={component}
          currentBuildId={latestDeploy?.build_id}
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

  return (
    <InstallComponentsList
      components={(result?.data ?? []).map(toListItem)}
      loading={isLoading}
      filtered={!!q || !!selectedTypes.length}
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
      filterActions={<ComponentTypeFilterDropdown options={FILTER_TYPES} />}
    />
  )
}
