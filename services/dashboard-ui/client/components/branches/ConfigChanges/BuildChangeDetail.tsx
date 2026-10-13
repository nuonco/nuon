import { Badge } from '@/components/common/Badge'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import {
  changeReasonBadgeTheme,
  changeReasonLabel,
} from '@/components/branches/WorkflowStepDetail/shared/format'
import type { TTemplateBuildChange } from './config-changes'

export const BuildChangeDetail = ({
  build,
}: {
  build: TTemplateBuildChange
}) => {
  const subject = build.commit?.message?.split('\n')[0]

  return (
    <span className="flex flex-col gap-3">
      <span className="flex flex-wrap items-center justify-between gap-3">
        <span className="flex flex-wrap items-center gap-3">
          {build.changeReason ? (
            <Badge size="sm" theme={changeReasonBadgeTheme(build.changeReason)}>
              {changeReasonLabel(build.changeReason)}
            </Badge>
          ) : null}
          <Status status={build.status} />
        </span>
        {build.href ? <Link href={build.href}>View build</Link> : null}
      </span>
      {build.commit ? (
        <span className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-1">
          <CommitLink sha={build.commit.sha} repo={build.commit.repo} />
          {subject ? (
            <Text variant="subtext" className="min-w-0 truncate">
              {subject}
            </Text>
          ) : null}
          {build.commit.author ? (
            <Text variant="subtext" theme="neutral">
              {build.commit.author}
            </Text>
          ) : null}
        </span>
      ) : null}
    </span>
  )
}
