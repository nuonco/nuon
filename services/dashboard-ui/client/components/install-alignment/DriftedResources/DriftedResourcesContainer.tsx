import { useSearchParams } from 'react-router'
import type { TCollectionView } from '@/components/common/CollectionViewToggle'
import { ImageConsumersContainer } from '@/components/install-components/InstallImagesList/ImageConsumers'
import { useInstall } from '@/hooks/use-install'
import { useInstallOverview } from '@/hooks/use-install-overview'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useResolveInstallDrift } from '@/hooks/use-resolve-install-drift'
import { useSurfaces } from '@/hooks/use-surfaces'
import type {
  TComponentType,
  TInstallAlignmentResource,
  TInstallAlignmentResourceKind,
} from '@/types'
import {
  branchRunHref,
  explanationEvidence,
  hasOngoingDeployment,
  isDrifted,
  resourcesOfKinds,
} from '../alignment-utils'
import { DriftPlanPanel } from '../DriftPlanPanel'
import { DriftResourceCard } from '../DriftResourceCard'
import { DriftedResources } from './DriftedResources'

export const useDriftCardProps = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { addPanel } = useSurfaces()
  const resolve = useResolveInstallDrift()
  const workflowHref = (workflowId?: string, stepId?: string) =>
    workflowId
      ? installLink({
          installId: install?.id,
          appId: install?.app_id,
          suffix: `/workflows/${workflowId}${stepId ? `?panel=${encodeURIComponent(stepId)}` : ''}`,
        })
      : undefined

  return (resource: TInstallAlignmentResource) => {
    const evidence = explanationEvidence(resource)
    return {
      resource,
      nameHref:
        resource.kind === 'component' || resource.kind === 'image'
          ? installLink({
              installId: install?.id,
              appId: install?.app_id,
              suffix: `/components/${resource.component_id}`,
            })
          : undefined,
      branchRunHref: branchRunHref({
        orgId: org?.id,
        appId: install?.app_id,
        commit: resource.last_attempt?.branch_run,
      }),
      explanationHref: workflowHref(evidence?.workflowId, evidence?.stepId),
      lastAttemptHref: workflowHref(resource.last_attempt?.workflow_id),
      deploymentHref: workflowHref(resource.open_deployment?.id),
      metadata:
        resource.kind === 'image' && resource.component_id ? (
          <ImageConsumersContainer componentId={resource.component_id} />
        ) : undefined,
      onViewPlan: resource.plan?.approval_id
        ? () => addPanel(<DriftPlanPanel resourceId={resource.id!} />)
        : undefined,
      onResolve: isDrifted(resource)
        ? () => resolve.mutate({ resource })
        : undefined,
      resolving:
        resolve.isPending && resolve.variables?.resource.id === resource.id,
      resolveError:
        resolve.variables?.resource.id === resource.id
          ? resolve.error
          : undefined,
    }
  }
}

export const DriftedResourcesContainer = ({
  kinds,
  noun,
  search,
  types,
  enabled,
  view,
}: {
  kinds: TInstallAlignmentResourceKind[]
  noun: string
  search?: string
  types?: TComponentType[]
  enabled?: string
  view?: TCollectionView
}) => {
  const [searchParams] = useSearchParams()
  const { data, error, isLoading, refetch } = useInstallOverview()
  const cardProps = useDriftCardProps()

  return (
    <DriftedResources
      resources={resourcesOfKinds(data?.alignment?.resources, kinds).filter(
        (resource) =>
          (resource.category !== 'torn-down' ||
            hasOngoingDeployment(resource)) &&
          (!types?.length ||
            types.includes(resource.component_type as TComponentType)) &&
          (enabled !== 'disabled' || resource.category === 'disabled') &&
          (enabled !== 'enabled' || resource.category !== 'disabled')
      )}
      loading={isLoading}
      error={error}
      unavailable={!!data?.alignment?.error}
      onRetry={() => void refetch()}
      noun={noun}
      search={search ?? searchParams.get('q') ?? undefined}
      view={view}
      renderCard={(resource) => (
        <DriftResourceCard {...cardProps(resource)} view={view} />
      )}
    />
  )
}
