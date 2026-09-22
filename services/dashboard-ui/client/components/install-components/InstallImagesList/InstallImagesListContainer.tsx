import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { BuildComponentButton } from '@/components/components/management/BuildComponent'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents } from '@/lib'
import type { TInstallComponent } from '@/types'
import { InstallImageSummary } from './InstallImageSummary'
import {
  InstallImagesList,
  type TInstallImageListItem,
} from './InstallImagesList'

const IMAGE_TYPES = 'external_image,docker_build'

export const InstallImagesListContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { appConfig } = useInstallAppConfig()
  const [searchParams] = useSearchParams()

  const q = searchParams.get('q') || undefined

  const { data: result, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-resource-images', org?.id, install?.id, q],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id,
        installId: install.id,
        limit: 100,
        offset: 0,
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

    return {
      id: componentId || (installComponent.id ?? ''),
      name: component?.name ?? 'Image',
      type: component?.type,
      status: installComponent.status_v2?.status ?? installComponent.status,
      buildAction: component ? (
        <BuildComponentButton
          component={component}
          size="sm"
          variant="secondary"
          redirectOnSuccess={false}
        >
          Build image
        </BuildComponentButton>
      ) : null,
      image: (
        <InstallImageSummary
          componentId={componentId}
          sourceRef={config?.external_image?.image_url}
        />
      ),
    }
  }

  return (
    <InstallImagesList
      images={(result?.data ?? []).map(toListItem)}
      loading={isLoading}
      filtered={!!q}
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
