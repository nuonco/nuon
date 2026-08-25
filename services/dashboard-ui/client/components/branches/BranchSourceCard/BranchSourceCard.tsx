import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import {
  BranchRunCommit,
  type IBranchRunCommit,
} from '@/components/branches/BranchRunCommit'
import type { TAppBranchConfig } from '@/types'

const SourceField = ({ label, value }: { label: string; value?: string }) => {
  if (!value) return null

  return (
    <div className="flex flex-col gap-0.5">
      <Text variant="label" theme="neutral">
        {label}
      </Text>
      <Text variant="subtext" className="break-all">
        {value}
      </Text>
    </div>
  )
}

export interface IBranchSourceCard {
  config?: TAppBranchConfig
  latestRun?: IBranchRunCommit
}

export const BranchSourceCard = ({ config, latestRun }: IBranchSourceCard) => {
  const connectedVCS = config?.connected_github_vcs_config
  const publicVCS = config?.public_git_vcs_config
  const vcs = connectedVCS ?? publicVCS
  const repoHref = vcs?.repo
    ? vcs.repo.startsWith('http')
      ? vcs.repo
      : `https://github.com/${vcs.repo}`
    : undefined

  return (
    <Card className="gap-4 p-4">
      <div className="flex items-center justify-between gap-2 flex-wrap">
        <div className="flex flex-col gap-0.5">
          <span className="flex items-center gap-1.5">
            <Icon variant="GitHub" size={14} />
            <Text weight="strong">Source</Text>
          </span>
          <Text variant="subtext" theme="neutral">
            {vcs
              ? connectedVCS
                ? 'Watched via GitHub connection'
                : 'Watched via public git repository'
              : 'No repository connected. Set a source when creating a deployment plan config.'}
          </Text>
        </div>
        {repoHref ? (
          <Text variant="subtext">
            <Link href={repoHref} isExternal>
              View on GitHub
            </Link>
          </Text>
        ) : null}
      </div>
      {vcs ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <SourceField label="Repository" value={vcs.repo} />
          <SourceField label="Branch" value={vcs.branch} />
          <SourceField label="Directory" value={vcs.directory} />
          <SourceField label="Path filter" value={vcs.path_filter} />
        </div>
      ) : null}
      {latestRun ? (
        <div className="flex flex-col gap-1.5 border-t pt-4">
          <Text variant="label" theme="neutral">
            Latest run
          </Text>
          <BranchRunCommit {...latestRun} />
        </div>
      ) : null}
    </Card>
  )
}
