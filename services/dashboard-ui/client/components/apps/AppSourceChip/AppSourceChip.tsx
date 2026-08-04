import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

export interface IAppSourceChip {
  connectHref: string
  isLoading?: boolean
  repo?: string
  repoHref?: string
}

export const AppSourceChip = ({
  connectHref,
  isLoading = false,
  repo,
  repoHref,
}: IAppSourceChip) => {
  if (isLoading) return null

  if (!repo) {
    return (
      <Link
        href={connectHref}
        className="flex items-center gap-1.5 w-fit text-sm"
      >
        <Icon variant="GitHub" size={14} />
        Connect repository
      </Link>
    )
  }

  return (
    <div className="flex items-center gap-2 w-fit border rounded-full px-2.5 py-1">
      <Icon variant="GitHub" size={14} />
      {repoHref ? (
        <Link
          href={repoHref}
          isExternal
          className="text-sm truncate max-w-72"
        >
          {repo}
        </Link>
      ) : (
        <Text variant="subtext" className="truncate max-w-72">
          {repo}
        </Text>
      )}
    </div>
  )
}
