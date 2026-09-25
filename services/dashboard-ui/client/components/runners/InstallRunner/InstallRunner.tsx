import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { ProcessCardComponent } from '@/components/runners/ProcessCard'

export type TInstallRunnerVariant = 'page' | 'embedded'

export interface IInstallRunner {
  actions?: ReactNode
  processes?: ReactNode
  processesLoading?: boolean
  recentJobs?: ReactNode
  runnerId?: string
  statusBanner?: ReactNode
  variant?: TInstallRunnerVariant
}

export const InstallRunner = ({
  actions,
  processes,
  processesLoading,
  recentJobs,
  runnerId,
  statusBanner,
  variant = 'page',
}: IInstallRunner) => {
  const isEmbedded = variant === 'embedded'

  return (
    <div className="flex flex-col gap-4 md:gap-6">
      {isEmbedded ? (
        <div className="flex flex-row flex-wrap items-center justify-between gap-4">
          <span className="min-w-0">
            <ID>{runnerId}</ID>
          </span>
          {actions}
        </div>
      ) : (
        <DetailHeader
          backLink={false}
          title="Install runner"
          id={runnerId}
          actions={actions}
        />
      )}

      {statusBanner}

      {isEmbedded ? null : <SectionHeader title="Processes" />}

      {processesLoading ? (
        <div className="@container">
          <div className="grid grid-cols-1 @5xl:grid-cols-2 gap-6 items-start">
            <ProcessCardComponent loading />
            <ProcessCardComponent loading />
          </div>
        </div>
      ) : processes ? (
        processes
      ) : (
        <Card>
          <EmptyState
            emptyTitle="No active processes"
            emptyMessage="No runner processes are currently active or offline."
            variant="table"
          />
        </Card>
      )}

      {recentJobs ? (
        <>
          <SectionHeader title="Recent jobs" />
          {recentJobs}
        </>
      ) : null}
    </div>
  )
}

export const InstallRunnerMissing = () => (
  <EmptyState
    emptyTitle="No runner"
    emptyMessage="This install does not have a runner yet."
    variant="diagram"
  />
)
