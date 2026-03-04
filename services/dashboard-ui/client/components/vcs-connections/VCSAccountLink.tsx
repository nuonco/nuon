import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TVCSConnection } from '@/types'

export const VCSAccountLink = ({
  vcs_connection,
}: {
  vcs_connection: TVCSConnection
}) => {
  return (
    <Link
      className="leading-none"
      href={`https://github.com/${vcs_connection?.github_account_name}`}
      isExternal
    >
      <Text variant="subtext" family="mono">
        {vcs_connection?.github_account_name ||
          vcs_connection?.github_account_id ||
          vcs_connection?.github_install_id}
      </Text>
    </Link>
  )
}
