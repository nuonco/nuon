import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RunningClock } from '@/components/workflows/RunningClock'
import type { TInstallRunbookRun } from '@/lib/ctl-api/installs/runbooks'

export interface ILatestRunbookRunCard {
  flush?: boolean
  href?: string
  isLoading?: boolean
  run?: TInstallRunbookRun
}

export const LatestRunbookRunCard = ({
  flush = false,
  href,
  isLoading,
  run,
}: ILatestRunbookRunCard) => {
  if (!isLoading && !run) {
    return (
      <Text variant="subtext" theme="neutral">
        No runs yet
      </Text>
    )
  }

  const Wrapper = flush ? 'div' : Card
  const workflow = run?.install_workflow
  const status = workflow?.status?.status ?? run?.status
  const statusDescription =
    run?.install_workflow?.status?.status_human_description ??
    run?.status_description

  return (
    <Wrapper className={flush ? 'flex flex-col gap-4' : undefined}>
      <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
        <LabeledStatus
          label="Status"
          loading={isLoading}
          statusProps={{ status }}
          tooltipProps={{
            tipContent: statusDescription,
            position: 'bottom',
          }}
        />
        <LabeledValue label="Started" loading={isLoading}>
          <Time variant="subtext" time={run?.created_at} format="relative" />
        </LabeledValue>
        <LabeledValue label="Duration" loading={isLoading}>
          {workflow?.finished ? (
            <Duration variant="subtext" nanoseconds={workflow.execution_time} />
          ) : workflow ? (
            <RunningClock
              label=""
              theme="default"
              startedAt={workflow.started_at || run?.created_at}
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )}
        </LabeledValue>
      </div>
      {href ? <Link href={href}>View run</Link> : null}
    </Wrapper>
  )
}
