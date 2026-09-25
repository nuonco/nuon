import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Time } from '@/components/common/Time'
import type { TInstallActionRun } from '@/types'

export interface ILatestActionRunCard {
  flush?: boolean
  href?: string
  isLoading?: boolean
  run?: TInstallActionRun
  trigger?: ReactNode
}

export const LatestActionRunCard = ({
  flush = false,
  href,
  isLoading,
  run,
  trigger,
}: ILatestActionRunCard) => {
  if (!isLoading && !run) {
    return (
      <EmptyState
        variant="table"
        size="sm"
        emptyTitle="No runs yet"
        emptyMessage="Runs appear here once this action is triggered."
      />
    )
  }

  const Wrapper = flush ? 'div' : Card

  return (
    <Wrapper className={flush ? 'flex flex-col gap-4' : undefined}>
      <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
        <LabeledStatus
          label="Status"
          loading={isLoading}
          statusProps={{ status: run?.status_v2?.status }}
          tooltipProps={{
            tipContent: run?.status_v2?.status_human_description,
            position: 'bottom',
          }}
        />
        <LabeledValue label="Started" loading={isLoading}>
          <Time variant="subtext" time={run?.created_at} format="relative" />
        </LabeledValue>
        <LabeledValue label="Duration" loading={isLoading}>
          <Duration variant="subtext" nanoseconds={run?.execution_time} />
        </LabeledValue>
        <LabeledValue label="Trigger" loading={isLoading}>
          {trigger}
        </LabeledValue>
      </div>
      {href ? <Link href={href}>View run</Link> : null}
    </Wrapper>
  )
}
