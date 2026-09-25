import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Time } from '@/components/common/Time'
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
      <EmptyState
        variant="table"
        size="sm"
        emptyTitle="No runs yet"
        emptyMessage="Runs appear here once this runbook is executed."
      />
    )
  }

  const Wrapper = flush ? 'div' : Card
  const status = run?.install_workflow?.status?.status ?? run?.status
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
          <Duration
            variant="subtext"
            beginTime={run?.created_at}
            endTime={run?.updated_at}
          />
        </LabeledValue>
      </div>
      {href ? <Link href={href}>View run</Link> : null}
    </Wrapper>
  )
}
