import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TVCSConnection } from '@/types'

export const VCSAccountLink = ({
  vcs_connection,
  href,
}: {
  vcs_connection: TVCSConnection
  href?: string
}) => {
  return (
    <Link
      className="leading-none"
      href={href ?? `https://github.com/${vcs_connection?.github_account_name}`}
      isExternal={!href}
    >
      <Text variant="subtext" family="mono">
        {vcs_connection?.github_account_name ||
          vcs_connection?.github_account_id ||
          vcs_connection?.github_install_id}
      </Text>
    </Link>
  )
}
