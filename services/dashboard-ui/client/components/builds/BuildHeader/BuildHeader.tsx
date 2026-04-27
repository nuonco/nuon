import { BackLink } from '@/components/common/BackLink'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ComponentType } from '@/components/components/ComponentType'
import { ComponentConfigContextTooltip } from '@/components/components/ComponentConfigContextTooltip'
import { CommitDetails } from '@/components/common/CommitDetails'
import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { CancelRunnerJobButton } from '@/components/runners/CancelRunnerJob'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import type { TApp, TBuild, TComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

interface IBuildHeader {
  component: TComponent
  build: TBuild
  app: TApp
}

export const BuildHeader = ({ component, build, app }: IBuildHeader) => {
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
            {build?.runner_job &&
            build?.status_v2?.status !== 'active' &&
            build?.status_v2?.status !== 'error' ? (
              <CancelRunnerJobButton
                jobType="build"
                runnerJob={build?.runner_job}
              />
            ) : null}
            {build?.queue_signal ? (
              <AdminDashboardLink
                path={`/queue-signals?search=${build.id}`}
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
    </header>
  )
}
