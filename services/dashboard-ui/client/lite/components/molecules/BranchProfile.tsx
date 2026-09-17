import type { HTMLAttributes } from 'react'
import type { TAppBranch } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { cn } from '@/utils/classnames'
import { Icon } from '../atoms/Icon'
import { Text } from '../atoms/Text'
import { ID } from './ID'

export type TBranchProfileVariant = 'full' | 'modeline'

export interface IBranchProfile
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  branch?: TAppBranch | null
  loading?: boolean
  variant?: TBranchProfileVariant
}

export const BranchProfile = ({
  branch,
  loading = false,
  variant = 'full',
  className,
  ...props
}: IBranchProfile) => {
  const name = branch?.name?.trim() || 'Branch unavailable'
  const config = branch ? latestBranchConfig(branch) : undefined
  const directory =
    config?.connected_github_vcs_config?.directory ??
    config?.public_git_vcs_config?.directory

  if (variant === 'modeline') {
    return (
      <div
        className={cn('flex min-w-0 items-center gap-1.5', className)}
        {...props}
      >
        <Icon
          variant="GitBranchIcon"
          size={14}
          className="shrink-0 text-tertiary"
          aria-hidden
        />
        <Text
          variant="caption"
          family="mono"
          weight="medium"
          loading={loading}
          loadingWidth={10}
          className="truncate leading-tight"
        >
          {name}
        </Text>
      </div>
    )
  }

  return (
    <div
      className={cn('flex min-w-0 items-center gap-2 text-left', className)}
      {...props}
    >
      <Icon
        variant="GitBranchIcon"
        size={16}
        className="shrink-0 text-tertiary"
        aria-hidden
      />
      <span className="flex min-w-0 flex-col">
        <span className="flex min-w-0 items-baseline gap-1.5">
          <Text
            variant="caption"
            family="mono"
            weight="semibold"
            loading={loading}
            loadingWidth={10}
            className="truncate leading-tight"
          >
            {name}
          </Text>
          {loading ? (
            <Text
              variant="caption"
              family="mono"
              color="tertiary"
              loading
              loadingWidth={7}
              className="leading-tight"
            />
          ) : directory ? (
            <Text
              variant="caption"
              family="mono"
              color="tertiary"
              className="truncate leading-tight"
            >
              {directory}
            </Text>
          ) : null}
        </span>
        <ID
          value={branch?.id ?? '—'}
          label="Branch ID"
          truncate
          copyable={false}
          loading={loading}
          loadingWidth={14}
        />
      </span>
    </div>
  )
}
