import { Badge } from '@/components/common/Badge'
import { WorkflowPanelLink } from '@/components/workflows/InstallWorkflowPanel'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { Text } from '@/components/common/Text'
import type { TDeploy } from '@/types'
import { useInstallLink } from '@/hooks/use-install-path'

interface IDeployTimeline {
  deploys: TDeploy[]
  pagination: { hasNext: boolean; offset: number; limit: number }
  orgId: string
  installId: string
  componentId: string
  componentName: string
  isLoading: boolean
  error: unknown
  openWorkflowPanel?: boolean
  variant?: 'deploy' | 'sync'
}

export const DeployTimeline = ({
  deploys,
  pagination,
  orgId,
  installId,
  componentId,
  componentName,
  isLoading,
  error,
  openWorkflowPanel,
  variant = 'deploy',
}: IDeployTimeline) => {
  const installLink = useInstallLink()
  const isSync = variant === 'sync'

  if (isLoading) {
    return <TimelineSkeleton />
  }

  if (error || deploys.length === 0) {
    return (
      <EmptyState
        variant="table"
        emptyTitle={isSync ? 'No syncs' : 'No deploys'}
        emptyMessage={
          isSync
            ? 'This image has not been synced yet.'
            : 'This component has not been deployed yet.'
        }
      />
    )
  }

  return (
    <Timeline<TDeploy>
      events={deploys}
      pagination={pagination}
      renderEvent={(deploy) => {
        return (
          <TimelineEvent
            key={deploy.id}
            caption={<ID>{deploy?.id}</ID>}
            createdAt={deploy?.created_at}
            status={deploy?.status}
            title={
              <span className="flex items-center gap-2">
                <DeployTimelineTitle
                  componentId={componentId}
                  componentName={componentName}
                  deploy={deploy}
                  installId={installId}
                  installLink={installLink}
                  isSync={isSync}
                  openWorkflowPanel={openWorkflowPanel}
                  orgId={orgId}
                />
                {deploy?.status_v2?.status === 'drifted' ? (
                  <Badge variant="code" size="sm">
                    drift scan
                  </Badge>
                ) : null}
              </span>
            }
            underline={
              <Text variant="label" theme="neutral">
                {isSync ? 'Synced' : 'Deployed'} by: {deploy?.created_by?.email}
              </Text>
            }
          />
        )
      }}
    />
  )
}

const DeployTimelineTitle = ({
  componentId,
  componentName,
  deploy,
  installId,
  installLink,
  isSync,
  openWorkflowPanel,
  orgId,
}: {
  componentId: string
  componentName: string
  deploy: TDeploy
  installId: string
  installLink: ReturnType<typeof useInstallLink>
  isSync: boolean
  openWorkflowPanel?: boolean
  orgId: string
}) => {
  const label = `${componentName} ${
    deploy.install_deploy_type === 'teardown'
      ? 'teardown'
      : isSync
        ? 'sync'
        : 'deploy'
  }`
  const workflowId = deploy.workflow_id || deploy.install_workflow_id

  if (openWorkflowPanel) {
    if (!workflowId) return <span>{label}</span>
    return (
      <WorkflowPanelLink variant="inline" workflowId={workflowId}>
        {label}
      </WorkflowPanelLink>
    )
  }

  return (
    <Link
      href={installLink({
        orgId,
        installId,
        suffix: `/components/${componentId}/deploys/${deploy.id}`,
      })}
      variant="inline"
    >
      {label}
    </Link>
  )
}
