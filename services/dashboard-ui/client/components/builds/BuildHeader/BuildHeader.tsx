import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toast } from '@/components/surfaces/Toast'
import { ComponentType } from '@/components/components/ComponentType'
import { ComponentConfigContextTooltip } from '@/components/components/ComponentConfigContextTooltip'
import { CommitDetails } from '@/components/common/CommitDetails'
import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { cancelComponentBuild } from '@/lib'
import type { TApp, TAPIError, TBuild, TComponent } from '@/types'
import { isImageBuild } from '@/utils/image-ref'
import { toSentenceCase } from '@/utils/string-utils'

interface IBuildHeader {
  component: TComponent
  build: TBuild
  app: TApp
}

export const BuildHeader = ({ component, build, app }: IBuildHeader) => {
  const { org } = useOrg()
  const { addToast } = useToast()
  const [hasBeenCanceled, setHasBeenCanceled] = useState(false)

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
    <header className="p-6 border-b flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <BackLink />
        <div className="flex items-center gap-6">
          <LabeledStatus
            label="Status"
            statusProps={{
              status: build?.status_v2?.status,
            }}
            tooltipProps={{
              tipContentClassName: 'w-fit',
              tipContent: (
                <Text nowrap variant="subtext">
                  {toSentenceCase(build?.status_v2?.status_human_description)}
                </Text>
              ),
              position: 'bottom',
            }}
          />
          <LabeledValue label="App">
            <Text variant="subtext">
              <Link href={`/${app?.org_id}/apps/${app?.id}`}>{app?.name}</Link>
            </Text>
          </LabeledValue>
          <LabeledValue label="Config">
            <ComponentConfigContextTooltip
              componentId={component?.id}
              configId={build?.component_config_connection?.id}
              appId={component?.app_id}
            >
              <Text variant="subtext">
                <Link
                  href={`/${app.org_id}/apps/${app.id}/components/${build?.component_id}`}
                >
                  {component?.name}
                </Link>
              </Text>
            </ComponentConfigContextTooltip>
          </LabeledValue>
          {build?.vcs_connection_commit ? (
            <LabeledValue label="Commit">
              <CommitDetails commit={build?.vcs_connection_commit} />
            </LabeledValue>
          ) : null}
        </div>
      </div>

      <div className="flex flex-col gap-1">
        <span className="flex items-center gap-2">
          <ComponentType type={component?.type} displayVariant="icon-only" />
          <Text variant="base" weight="strong">
            {component?.name} build
          </Text>
          {build?.no_op ? (
            <Badge variant="code" size="sm" theme="neutral">
              no-op
            </Badge>
          ) : null}
        </span>
        <ID>{build?.id}</ID>
        <div className="flex items-center justify-between mt-1">
          <div className="flex gap-8 items-center">
            <Text theme="info" flex className="gap-1">
              <Icon variant="CalendarBlankIcon" />
              <Time variant="subtext" time={build.created_at} />
            </Text>
            <Text theme="info" flex className="gap-1">
              <Icon variant="TimerIcon" />
              <Duration
                variant="subtext"
                beginTime={build.created_at}
                endTime={build.updated_at}
              />
            </Text>
          </div>
          <div className="flex items-center gap-4">
            {isCancelable ? (
              <Button
                variant="danger"
                disabled={isCanceling || hasBeenCanceled}
                onClick={() => cancelBuild()}
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
            {build?.runner_job ? (
              <RunnerJobPlanButton
                buttonText="Build plan"
                runnerJobId={build?.runner_job?.id}
              />
            ) : null}
          </div>
        </div>
      </div>

      {isImageBuild(build) ? <BuildImageSourceDetails build={build} /> : null}
    </header>
  )
}

const BuildImageSourceDetails = ({ build }: { build: TBuild }) => {
  return (
    <div className="flex flex-col gap-3 pt-4 border-t">
      <Text variant="subtext" weight="strong">
        Image source
      </Text>
      <div className="grid gap-4 md:grid-cols-2">
        {build.source_ref ? (
          <LabeledValue label="Source ref">
            <Text variant="subtext" family="mono" className="break-all">
              {build.source_ref}
            </Text>
          </LabeledValue>
        ) : null}
        {build.resolved_tag ? (
          <LabeledValue label="Resolved tag">
            <Text variant="subtext" family="mono">
              {build.resolved_tag}
            </Text>
          </LabeledValue>
        ) : null}
        {build.source_digest ? (
          <LabeledValue label="Digest" className="md:col-span-2">
            <ClickToCopy>
              <Text variant="subtext" family="mono" className="break-all">
                {build.source_digest}
              </Text>
            </ClickToCopy>
          </LabeledValue>
        ) : null}
        {build.source_media_type ? (
          <LabeledValue label="Media type" className="md:col-span-2">
            <Text variant="subtext" family="mono" className="break-all">
              {build.source_media_type}
            </Text>
          </LabeledValue>
        ) : null}
        {build.resolved_at ? (
          <LabeledValue label="Resolved">
            <Time variant="subtext" time={build.resolved_at} />
          </LabeledValue>
        ) : null}
      </div>
    </div>
  )
}
