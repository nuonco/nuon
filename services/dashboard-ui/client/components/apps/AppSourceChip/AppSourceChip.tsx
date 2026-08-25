import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

export interface IAppSourceChip {
  isLoading?: boolean
  repo?: string
  repoHref?: string
}

export const AppSourceChip = ({
  isLoading = false,
  repo,
  repoHref,
}: IAppSourceChip) => {
  if (isLoading || !repo) return null

  return (
    <div className="flex items-center gap-2 w-fit border rounded-full px-2.5 py-1">
      <Icon variant="GitHub" size={14} />
      {repoHref ? (
        <Link
          href={repoHref}
          isExternal
          textVariant="body"
          className="truncate max-w-72"
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
