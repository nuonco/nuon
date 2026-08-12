import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'

export interface IBranchVcsBadges {
  repo?: string
  branch?: string
  size?: 'sm' | 'md'
}

export const BranchVcsBadges = ({
  repo,
  branch,
  size = 'sm',
}: IBranchVcsBadges) => {
  if (!repo && !branch) return null

  return (
    <>
      {repo ? (
        <Badge size={size} theme="default">
          <Icon variant="GitHub" size={13} />
          {repo}
        </Badge>
      ) : null}
      {branch ? (
        <Badge size={size} theme="default">
          <Icon variant="GitBranchIcon" size={13} />
          {branch}
        </Badge>
      ) : null}
    </>
  )
}
