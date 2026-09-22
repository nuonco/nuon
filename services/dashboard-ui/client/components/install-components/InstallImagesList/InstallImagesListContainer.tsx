import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { ResourceComponentActions } from '@/components/install-components/ResourceComponentActions'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents } from '@/lib'
import type { TInstallComponent } from '@/types'
import { InstallImageSummaryContainer } from './InstallImageSummary'
import {
  InstallImagesList,
  type TInstallImageListItem,
} from './InstallImagesList'

const IMAGE_TYPES = 'external_image,docker_build'
const LIMIT = 10

export const InstallImagesListContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { appConfig } = useInstallAppConfig()
  const [searchParams] = useSearchParams()

  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined

  const { data: result, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-resource-images', org?.id, install?.id, offset, q],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q,
        types: IMAGE_TYPES,
      }),
    refetchInterval: 20000,
    enabled: !!org?.id && !!install?.id,
  })

  const toListItem = (
    installComponent: TInstallComponent
  ): TInstallImageListItem => {
    const component = installComponent.component
    const componentId = component?.id ?? installComponent.component_id ?? ''
    const config = appConfig?.component_config_connections?.find(
      (connection) => connection.component_id === componentId
    )
    const latestDeploy = installComponent.install_deploys?.[0]

    return {
      id: componentId || (installComponent.id ?? ''),
      name: component?.name ?? 'Image',
      type: component?.type,
      status: installComponent.status_v2?.status ?? installComponent.status,
      actions: component ? (
        <ResourceComponentActions
          component={component}
          currentBuildId={latestDeploy?.build_id}
          currentDeployStatus={
            latestDeploy?.status_v2?.status ?? latestDeploy?.status
          }
          variant="image"
        />
      ) : null,
      image: componentId ? (
        <InstallImageSummaryContainer
          componentId={componentId}
          sourceRef={config?.external_image?.image_url}
        />
      ) : null,
    }
  }

  return (
    <InstallImagesList
      images={(result?.data ?? []).map(toListItem)}
      loading={isLoading}
      filtered={!!q}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
      search={
        <DebouncedSearchInput
          className="w-full md:w-fit"
          labelClassName="w-full md:w-fit"
          placeholder="Search by name or ID..."
        />
      }
    />
  )
}
