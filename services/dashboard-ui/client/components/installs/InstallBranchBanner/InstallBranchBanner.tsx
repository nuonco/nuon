import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TAPIError } from '@/types'

export interface IInstallBranchBanner {
  branchHref?: string
  branchName?: string
  runsBehind?: number | null
  error?: TAPIError | null
  onRetry?: () => void
  onViewRuns?: () => void
  reconcileAction?: ReactNode
}

export const InstallBranchBanner = ({
  branchHref,
  branchName,
  runsBehind,
  error,
  onRetry,
  onViewRuns,
  reconcileAction,
}: IInstallBranchBanner) => {
  if (error) {
    return (
      <Banner theme="error" className="!py-2" data-testid="branch-banner-error">
        <div className="flex flex-wrap justify-between items-center gap-3 w-full">
          <Text variant="subtext">
            Unable to load branch tracking for this install.
          </Text>
          {onRetry ? (
            <Button size="sm" onClick={onRetry}>
              Try again
            </Button>
          ) : null}
        </div>
      </Banner>
    )
  }

  if (!runsBehind || runsBehind < 1) return null

  const branch = (
    <Text as="span" family="mono" variant="subtext">
      {branchName ?? 'its app branch'}
    </Text>
  )

  return (
    <Banner
      theme="info"
      className="!py-2 dark:!text-blue-300"
      data-testid="branch-banner"
    >
      <div className="flex flex-wrap justify-between items-center gap-3 w-full">
        <Text variant="subtext">
          This install is{' '}
          <strong>
            {runsBehind} app branch {runsBehind === 1 ? 'run' : 'runs'} behind
          </strong>{' '}
          {branchHref ? <Link href={branchHref}>{branch}</Link> : branch}.
        </Text>
        <div className="flex items-center gap-2 shrink-0">
          {onViewRuns ? (
            <Button size="sm" onClick={onViewRuns}>
              View runs
            </Button>
          ) : null}
          {reconcileAction}
        </div>
      </div>
    </Banner>
  )
}
