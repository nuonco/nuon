import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { VCSAccountLink } from '@/components/vcs-connections/VCSAccountLink'
import type { TVCSConnection } from '@/types'

interface IGitHubAccountSection {
  vcs_connection: TVCSConnection
}

export const GitHubAccountSection = ({
  vcs_connection,
}: IGitHubAccountSection) => (
  <div className="flex flex-col gap-6">
    {/* GitHub Account Info */}
    <div className="flex flex-col gap-4">
      <Text variant="body" weight="strong">
        GitHub account
      </Text>
      <div className="grid grid-cols-3 gap-4">
        <LabeledValue label="Account name">
          <VCSAccountLink vcs_connection={vcs_connection} />
        </LabeledValue>

        <LabeledValue label="Account ID">
          <ID theme="default">{vcs_connection?.github_account_id || '—'}</ID>
        </LabeledValue>

        <LabeledValue label="Install ID">
          {vcs_connection?.github_install_id ? (
            <a
              href={`https://github.com/organizations/${vcs_connection.github_account_name}/settings/installations/${vcs_connection.github_install_id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 hover:underline"
            >
              <ID theme="default">{vcs_connection.github_install_id}</ID>
              <Icon variant="ArrowSquareOutIcon" size={12} className="text-cool-grey-400" />
            </a>
          ) : (
            <ID theme="default">—</ID>
          )}
        </LabeledValue>
      </div>
    </div>

    {/* Connection Details */}
    <div className="flex flex-col gap-4">
      <Text variant="body" weight="strong">
        Connection details
      </Text>
      <div className="grid grid-cols-3 gap-4">
        <LabeledValue label="Connection ID">
          <ID theme="default">{vcs_connection?.id || '—'}</ID>
        </LabeledValue>

        <LabeledValue label="Created">
          <Text variant="subtext">
            {vcs_connection?.created_at ? (
              <Time
                variant="subtext"
                time={vcs_connection?.created_at}
                format="relative"
              />
            ) : (
              <Icon variant="MinusIcon" />
            )}
          </Text>
        </LabeledValue>

        <LabeledValue label="Last updated">
          <Text variant="subtext">
            {vcs_connection?.updated_at ? (
              <Time
                variant="subtext"
                time={vcs_connection?.updated_at}
                format="relative"
              />
            ) : (
              <Icon variant="MinusIcon" />
            )}
          </Text>
        </LabeledValue>
      </div>
    </div>

    {/* Related Commits (if available) */}
    {vcs_connection?.vcs_connection_commit?.length ? (
      <div className="flex flex-col gap-4">
        <Text variant="body" weight="strong">
          Related commits
        </Text>
        <LabeledValue label="Total commits">
          <Text variant="subtext">
            {vcs_connection.vcs_connection_commit.length}
          </Text>
        </LabeledValue>
      </div>
    ) : null}
  </div>
)
