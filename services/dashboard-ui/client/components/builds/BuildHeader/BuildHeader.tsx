import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toast } from '@/components/surfaces/Toast'
import { ComponentType } from '@/components/components/ComponentType'
import { ComponentConfigContextTooltip } from '@/components/components/ComponentConfigContextTooltip'
import { CommitDetails } from '@/components/common/CommitDetails'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { CancelBuildModal } from '@/components/builds/CancelBuild'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { cancelComponentBuild } from '@/lib'
import type { TApp, TAPIError, TBuild, TComponent } from '@/types'

interface IBuildHeader {
  component: TComponent
  build: TBuild
  app: TApp
}

export const BuildHeader = ({ component, build, app }: IBuildHeader) => {
  const { org } = useOrg()
  const { addToast } = useToast()
  const { addModal } = useSurfaces()
  const [hasBeenCanceled, setHasBeenCanceled] = useState(false)
  const executionJobId = build?.build_runner_job_id ?? build?.runner_job?.id

  const { mutate: cancelBuild, isPending: isCanceling } = useMutation<
    unknown,
    TAPIError
  >({
    mutationFn: () =>
      cancelComponentBuild({
        orgId: org.id,
        appId: app.id,
        componentId: component.id,
        buildId: build.id,
      }),
    onSuccess: () => {
      setHasBeenCanceled(true)
      addToast(
        <Toast heading="Build cancelled." theme="success">
          <Text>Successfully cancelled the build.</Text>
        </Toast>,
      )
    },
    onError: (err: { error?: string }) => {
      addToast(
        <Toast heading="Cancel build failed." theme="error">
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>,
      )
    },
  })

  const isCancelable =
    build?.queue_signal &&
    build?.status_v2?.status !== 'active' &&
    build?.status_v2?.status !== 'error' &&
    build?.status_v2?.status !== 'cancelled'

  return (
    <DetailHeader
      icon={<ComponentType type={component?.type} displayVariant="icon-only" />}
      title={`${component?.name} build`}
      status={
        build?.no_op ? (
          <Badge variant="code" size="sm" theme="neutral">
            no-op
          </Badge>
        ) : null
      }
      id={build?.id}
      identity={
        <Time
          time={build?.created_at}
          format="relative"
          variant="subtext"
          theme="info"
        />
      }
      actions={
        <>
          {isCancelable ? (
            <Button
              variant="danger"
              disabled={isCanceling || hasBeenCanceled}
              onClick={() =>
                addModal(
                  <CancelBuildModal
                    componentName={component?.name}
                    onConfirm={() => cancelBuild()}
                  />
                )
              }
            >
              {isCanceling ? (
                <span className="flex items-center gap-2">
                  <Icon variant="Loading" className="animate-spin" />
                  Canceling
                </span>
              ) : hasBeenCanceled ? (
                'Canceled'
              ) : (
                'Cancel build'
              )}
            </Button>
          ) : null}
          {build?.queue_signal ? (
            <AdminDashboardLink
              path={`/queues/${build.queue_signal.queue_id}/signals/${build.queue_signal.id}`}
              label="View signal"
            />
          ) : null}
          {executionJobId ? (
            <>
              <RunnerJobPlanButton
                buttonText="Build plan"
                runnerJobId={executionJobId}
              />
              <Button
                href={`/${org?.id}/runner/jobs/${executionJobId}`}
                variant="secondary"
              >
                View execution
              </Button>
            </>
          ) : null}
        </>
      }
      metadata={
        <>
          <LabeledStatus
            label="Status"
            statusProps={{
              status: build?.status_v2?.status,
            }}
            tooltipProps={{
              tipContentClassName: 'w-fit',
              tipContent: (
                <Text nowrap variant="subtext">
                  {build?.status_v2?.status_human_description}
                </Text>
              ),
              position: 'bottom',
            }}
          />
          <LabeledValue label="Duration">
            <Duration
              variant="subtext"
              beginTime={build?.created_at}
              endTime={build?.updated_at}
            />
          </LabeledValue>
          <LabeledValue label="App">
            <Link href={`/${app?.org_id}/apps/${app?.id}`}>{app?.name}</Link>
          </LabeledValue>
          <LabeledValue label="Config">
            <ComponentConfigContextTooltip
              componentId={component?.id}
              configId={build?.component_config_connection?.id}
              appId={component?.app_id}
            >
              <Link
                href={`/${app?.org_id}/apps/${app?.id}/components/${build?.component_id}`}
              >
                {component?.name}
              </Link>
            </ComponentConfigContextTooltip>
          </LabeledValue>
          {build?.vcs_connection_commit ? (
            <LabeledValue label="Commit">
              <CommitDetails commit={build?.vcs_connection_commit} />
            </LabeledValue>
          ) : null}
          {build?.app_branch_id && build?.app_branch_run_id ? (
            <LabeledValue label="Branch run">
              <Text variant="subtext" flex className="gap-1 items-center">
                <Icon variant="GitBranchIcon" size={14} />
                <Link
                  href={`/${app?.org_id}/apps/${app?.id}/branches/${build.app_branch_id}/runs/${build.app_branch_run_id}`}
                  variant="inline"
                >
                  View run
                </Link>
              </Text>
            </LabeledValue>
          ) : null}
        </>
      }
    />
  )
}
