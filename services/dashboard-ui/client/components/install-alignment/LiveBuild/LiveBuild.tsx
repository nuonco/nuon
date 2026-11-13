import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TInstallAlignmentBuild } from '@/types'
import { buildRef } from '../alignment-utils'

export const BuildRef = ({
  build,
  href,
}: {
  build?: TInstallAlignmentBuild
  href?: string
}) => {
  const ref = buildRef(build)
  if (!ref) {
    return (
      <Text as="span" variant="subtext" theme="neutral">
        Not deployed
      </Text>
    )
  }
  return href ? (
    <Link href={href} variant="inline" className="!font-mono text-sm">
      {ref}
    </Link>
  ) : (
    <Text as="span" family="mono" variant="subtext">
      {ref}
    </Text>
  )
}

export interface ILiveBuild {
  appliedAt?: string
  build?: TInstallAlignmentBuild
  buildHref?: string
  deployHref?: string
  variant?: 'deploy' | 'sync'
}

export const LiveBuild = ({
  appliedAt,
  build,
  buildHref,
  deployHref,
  variant = 'deploy',
}: ILiveBuild) => (
  <div
    className="flex flex-wrap items-center gap-x-3 gap-y-1"
    data-testid="live-build"
  >
    <div className="flex items-center gap-2">
      <Text as="span" variant="subtext" theme="neutral">
        Live
      </Text>
      <BuildRef build={build} href={buildHref} />
    </div>
    {appliedAt ? (
      <div className="flex items-center gap-1">
        <Text as="span" variant="subtext" theme="neutral">
          {variant === 'sync' ? 'Synced' : 'Applied'}
        </Text>
        <Time
          time={appliedAt}
          format="relative"
          variant="subtext"
          theme="neutral"
        />
      </div>
    ) : null}
    {deployHref ? (
      <Link href={deployHref}>
        {variant === 'sync' ? 'View sync' : 'View deployment'}
      </Link>
    ) : null}
  </div>
)
