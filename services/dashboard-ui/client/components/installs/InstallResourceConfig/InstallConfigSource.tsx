import { Badge } from '@/components/common/Badge'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TAppBranchRun } from '@/types'
import { ConfigSection } from './ConfigLayout'

export interface IInstallConfigSource {
  behind?: boolean
  isLoading?: boolean
  run?: TAppBranchRun
  runHref?: string
}

export const InstallConfigBehindBadge = () => (
  <Tooltip
    position="top"
    tipContent={
      <Text variant="subtext">
        Still on an older app config than this install.
      </Text>
    }
  >
    <Badge size="sm" theme="warn">
      Behind
    </Badge>
  </Tooltip>
)

export const InstallConfigSource = ({
  behind,
  isLoading,
  run,
  runHref,
}: IInstallConfigSource) => {
  const commit = run?.vcs_connection_commit
  const sha = commit?.sha || run?.head_sha

  return (
    <ConfigSection
      title="Source"
      actions={behind ? <InstallConfigBehindBadge /> : null}
    >
      <div className="rounded-md border px-4 py-3">
        {isLoading && !run ? (
          <Text variant="subtext" loading loadingWidth={32} />
        ) : run ? (
          <BranchRunCommit
            author={commit?.author_name}
            avatarUrl={commit?.author_avatar_url}
            createdAt={commit?.created_at ?? run.created_at}
            href={runHref}
            message={commit?.message?.split('\n')[0]}
            sha={sha}
            showStatus={!!run.status}
            status={run.status}
          />
        ) : (
          <Text variant="subtext" theme="neutral">
            No branch run is recorded for this app config.
          </Text>
        )}
      </div>
    </ConfigSection>
  )
}
