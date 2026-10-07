import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import {
  getInstallComponents,
  getInstallHealthTimeline,
  getInstallResources,
} from '@/lib'
import type { TAPIError } from '@/types'
import { InstallHealth } from './InstallHealth'
import { InstallHealthResources } from './InstallHealthResources'

export const InstallHealthContainer = () => {
  const { install } = useInstall()
  const { org } = useOrg()
  const installLink = useInstallLink()
  // The existing timeline endpoint supplies access errors and action checks.
  // Request its smallest window without rendering history.
  const { data, error } = useQuery({
    queryKey: ['install-health-timeline', org?.id, install?.id, 1],
    queryFn: () =>
      getInstallHealthTimeline({
        orgId: org!.id,
        installId: install!.id,
        days: 1,
      }),
    enabled: !!org?.id && !!install?.id,
    placeholderData: (previous, query) =>
      query?.queryKey[1] === org?.id && query?.queryKey[2] === install?.id
        ? keepPreviousData(previous)
        : undefined,
    refetchInterval: 15000,
  })

  const {
    data: resources,
    isPending: resourcesPending,
    error: resourcesError,
  } = useQuery({
    queryKey: ['install-resources', org?.id, install?.id],
    queryFn: () =>
      getInstallResources({ orgId: org!.id, installId: install!.id }),
    enabled: !!org?.id && !!install?.id,
    placeholderData: (previous, query) =>
      query?.queryKey[1] === org?.id && query?.queryKey[2] === install?.id
        ? keepPreviousData(previous)
        : undefined,
    refetchInterval: 15000,
  })
  const { data: components } = useQuery({
    queryKey: ['install-components-for-resources', org?.id, install?.id],
    queryFn: () =>
      getInstallComponents({
        orgId: org!.id,
        installId: install!.id,
        limit: 100,
        offset: 0,
      }),
    enabled: !!org?.id && !!install?.id,
  })
  const componentNames: Record<string, string> = {}
  components?.data?.forEach((component) => {
    if (component?.id)
      componentNames[component.id] = component.component?.name || component.id
  })

  return (
    <InstallHealth
      clusterAccessError={data?.cluster_access_error}
      healthchecks={data?.healthchecks}
      getHealthcheckHref={(check) =>
        check.workflow_id
          ? installLink({
              installId: install?.id,
              appId: install?.app_id,
              suffix: `/deployments/${check.workflow_id}`,
            })
          : undefined
      }
      error={
        error
          ? (error as TAPIError).error ||
            'Health checks and cluster access status failed to load.'
          : undefined
      }
      resources={
        <InstallHealthResources
          resources={resources}
          componentNames={componentNames}
          isLoading={resourcesPending}
          clusterAccessError={data?.cluster_access_error}
          error={
            resourcesError
              ? (resourcesError as TAPIError).error ||
                'Resource health failed to load.'
              : undefined
          }
        />
      }
    />
  )
}
