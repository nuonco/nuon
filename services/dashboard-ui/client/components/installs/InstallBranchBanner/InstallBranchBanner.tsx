import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TAPIError } from '@/types'

const compactBanner =
  '!py-2 items-center [&>div:first-child]:!mt-0 [&>div:first-child]:self-center'

export interface IInstallBranchBannerReconcile {
  status: 'behind' | 'blocked' | 'in_progress'
  pinnedVersion?: number
  targetVersion?: number
  workflowHref?: string
  action?: ReactNode
}

export interface IInstallBranchBanner {
  branchHref?: string
  branchName?: string
  runsBehind?: number | null
  reconcile?: IInstallBranchBannerReconcile
  error?: TAPIError | null
  onRetry?: () => void
  onViewRuns?: () => void
}

export const InstallBranchBanner = ({
  branchHref,
  branchName,
  runsBehind,
  reconcile,
  error,
  onRetry,
  onViewRuns,
}: IInstallBranchBanner) => {
  if (error) {
    return (
      <Banner
        theme="error"
        className={compactBanner}
        data-testid="branch-banner-error"
      >
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

  if (!reconcile && (!runsBehind || runsBehind < 1)) return null

  const branch = (
    <Text
      as="span"
      family="mono"
      variant="subtext"
      className="!leading-[inherit] align-baseline"
    >
      {branchName ?? 'its app branch'}
    </Text>
  )
  const branchLink = branchHref ? (
    <Link href={branchHref} variant="inline">
      {branch}
    </Link>
  ) : (
    branch
  )

  if (reconcile) {
    const { pinnedVersion, targetVersion } = reconcile
    return (
      <Banner
        theme="neutral"
        className={compactBanner}
        data-testid="branch-banner"
      >
        <div className="flex flex-wrap justify-between items-center gap-3 w-full">
          <Text variant="subtext">
            {reconcile.status === 'in_progress' ? (
              <>
                Reconciling this install onto {branchLink}
                {targetVersion ? <> config v{targetVersion}</> : null}.{' '}
                {reconcile.workflowHref ? (
                  <Link href={reconcile.workflowHref} variant="inline">
                    View workflow
                  </Link>
                ) : null}
              </>
            ) : pinnedVersion && targetVersion ? (
              <>
                This install is on <strong>config v{pinnedVersion}</strong>.{' '}
                {branchLink} is on config v{targetVersion}.
              </>
            ) : (
              <>
                This install is <strong>behind</strong> {branchLink}.
              </>
            )}
          </Text>
          <div className="flex items-center gap-2">
            {onViewRuns && runsBehind && runsBehind > 0 ? (
              <Button onClick={onViewRuns}>View branch runs</Button>
            ) : null}
            {reconcile.action}
          </div>
        </div>
      </Banner>
    )
  }

  return (
    <Banner
      theme="info"
      className={`${compactBanner} dark:!text-blue-300`}
      data-testid="branch-banner"
    >
      <div className="flex flex-wrap justify-between items-center gap-3 w-full">
        <Text variant="subtext">
          This install is{' '}
          <strong>
            {runsBehind} app branch {runsBehind === 1 ? 'run' : 'runs'} behind
          </strong>{' '}
          {branchLink}.
        </Text>
        {onViewRuns ? (
          <Button onClick={onViewRuns}>View branch runs</Button>
        ) : null}
      </div>
    </Banner>
  )
}
