import { useSearchParams } from 'react-router'
import type { TCollectionView } from '@/components/common/CollectionViewToggle'
import { ImageConsumersContainer } from '@/components/install-components/InstallImagesList/ImageConsumers'
import { useInstall } from '@/hooks/use-install'
import { useInstallOverview } from '@/hooks/use-install-overview'
import { useInstallLink } from '@/hooks/use-install-path'
import { useResolveInstallDrift } from '@/hooks/use-resolve-install-drift'
import { useSurfaces } from '@/hooks/use-surfaces'
import type {
  TComponentType,
  TInstallAlignmentResource,
  TInstallAlignmentResourceKind,
} from '@/types'
import {
  explanationEvidence,
  hasOngoingDeployment,
  isDrifted,
  resourcesOfKinds,
} from '../alignment-utils'
import { DriftPlanPanel } from '../DriftPlanPanel'
import { DriftResourceCard } from '../DriftResourceCard'
import { useOriginHref } from '../LiveBuild'
import { DriftedResources } from './DriftedResources'

export const useDriftCardProps = () => {
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { addPanel } = useSurfaces()
  const resolve = useResolveInstallDrift()
  const originHref = useOriginHref()
  const { data: overview } = useInstallOverview()
  const repo = overview?.branch_tracking?.repo
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
    const nameHref =
      resource.kind === 'component' || resource.kind === 'image'
        ? installLink({
            installId: install?.id,
            appId: install?.app_id,
            suffix: `/components/${resource.component_id}`,
          })
        : undefined
    return {
      resource,
      nameHref,
      originHref,
      repo,
      historyHref: nameHref ? `${nameHref}/deploys` : undefined,
      explanationHref: workflowHref(evidence?.workflowId, evidence?.stepId),
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
  const originHref = useOriginHref()

  return (
    <DriftedResources
      originHref={originHref}
      repo={data?.branch_tracking?.repo}
      selectedCommit={data?.branch_tracking?.selected_commit}
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
