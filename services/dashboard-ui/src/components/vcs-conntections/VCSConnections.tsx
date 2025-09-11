import { FaGithub } from 'react-icons/fa'
import { Text } from '@/components/Typography'
import type { TVCSConnection } from '@/types'
import { RemoveVCSConnection } from './RemoveVCSConnection'

export const VCSConnections = ({
  vcsConnections,
}: {
  vcsConnections: TVCSConnection[]
}) => {
  return (
    <>
      {vcsConnections?.length &&
        vcsConnections?.map((vcs) => (
          <Text
            key={vcs?.id}
            className="flex gap-2 py-2 items-center font-mono text-sm text-cool-grey-600 dark:text-cool-grey-500 w-full"
          >
            <FaGithub className="text-lg" />
            {vcs?.github_account_name || vcs?.github_install_id}
            <RemoveVCSConnection connection={vcs} />
          </Text>
        ))}
    </>
  )
}
