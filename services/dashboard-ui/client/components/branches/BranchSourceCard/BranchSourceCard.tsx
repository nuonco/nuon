import { Card } from '@/components/common/Card'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import {
  BranchRunCommit,
  type IBranchRunCommit,
} from '@/components/branches/BranchRunCommit'
import type { TAppBranchConfig } from '@/types'

const runCadence = (config?: TAppBranchConfig) => {
  const runConfig = config?.run_config
  const mode =
    !runConfig?.mode || runConfig.mode === 'all' ? 'push' : runConfig.mode

  switch (mode) {
    case 'on_tag_prefix':
      return {
        mode,
        description: `Tags matching ${runConfig?.tag_prefix ?? 'the configured prefix'}`,
      }
    case 'on_github_label':
      return {
        mode,
        description: `Merged PRs labeled ${runConfig?.github_label ?? 'with the configured label'}`,
      }
    case 'manual_only':
      return { mode, description: 'Manual updates' }
    default:
      return { mode: 'push', description: 'Every push' }
  }
}

const SourceField = ({ label, value }: { label: string; value?: string }) => {
  if (!value) return null

  return (
    <LabeledValue label={label}>
      <Text variant="subtext" className="break-all">
        {value}
      </Text>
    </LabeledValue>
  )
}

export interface IBranchSourceCard {
  config?: TAppBranchConfig
  latestRun?: IBranchRunCommit
  onEdit?: () => void
}

export const BranchSourceCard = ({
  config,
  latestRun,
  onEdit,
}: IBranchSourceCard) => {
  const connectedVCS = config?.connected_github_vcs_config
  const publicVCS = config?.public_git_vcs_config
  const vcs = connectedVCS ?? publicVCS
  const repoHref = vcs?.repo
    ? (vcs.repo.startsWith('http')
        ? vcs.repo
        : `https://github.com/${vcs.repo}`
      ).replace(/\.git\/?$/, '')
    : undefined
  const sourceHref =
    repoHref && vcs?.branch ? `${repoHref}/tree/${vcs.branch}` : repoHref
  const cadence = runCadence(config)

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
        <div className="flex items-center gap-3">
          {sourceHref ? (
            <Link href={sourceHref} isExternal>
              View on GitHub
            </Link>
          ) : null}
          {onEdit ? (
            <Button variant="secondary" onClick={onEdit}>
              <Icon variant="PencilSimpleLineIcon" size={16} />
              Edit source
            </Button>
          ) : null}
        </div>
      </div>
      {vcs ? (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <SourceField label="Repository" value={vcs.repo} />
          <SourceField label="Branch" value={vcs.branch} />
          <SourceField label="Directory" value={vcs.directory} />
          <LabeledValue label="Run cadence">
            <span className="flex items-center gap-2 min-w-0">
              <Text variant="subtext" className="truncate">
                {cadence.description}
              </Text>
              <LabelBadge
                labelKey="mode"
                labelValue={cadence.mode}
                size="xs"
                theme={cadence.mode === 'push' ? 'default' : 'brand'}
              />
            </span>
          </LabeledValue>
        </div>
      ) : null}
      {latestRun ? (
        <LabeledValue label="Latest run" className="border-t pt-4">
          <BranchRunCommit {...latestRun} />
        </LabeledValue>
      ) : null}
    </Card>
  )
}
