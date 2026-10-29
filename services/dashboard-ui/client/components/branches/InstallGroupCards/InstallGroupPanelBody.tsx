import { useState, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { CompositeError } from '@/components/common/CompositeError'
import { Link } from '@/components/common/Link'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { humanize } from '@/utils/string-utils'
import type {
  TTrackGroup,
  TTrackInstall,
} from '@/components/branches/BranchOverview/RolloutTrack'
import { CommitRange, type ICommitRange } from './CommitRange'
import { GroupLabels } from './GroupLabels'
import { GroupApproveAllButton } from './InstallApprovalActions'
import { InstallRolloutCard } from './InstallRolloutCard'
import { InstallStatusCounts } from './InstallStatusCounts'
import { filterInstalls, withQueuedInstalls } from './install-status'

export const InstallGroupPanelBody = ({
  group,
  commit,
  approval,
  onSelectInstall,
  installLayout = 'card',
}: {
  group: TTrackGroup
  commit?: ICommitRange
  approval?: ReactNode
  onSelectInstall?: (install: TTrackInstall) => void
  installLayout?: 'card' | 'row'
}) => {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<string>()
  const counted = withQueuedInstalls(group.installs, group.plannedCount)
  const statuses = [...new Set(group.installs.map((install) => install.status))]
  const installs = filterInstalls(group.installs, { query: search, status })
  const failedInstallId = (
    group.error?.data as unknown as { install_id?: string } | undefined
  )?.install_id
  const failedInstallHref = group.installs.find(
    (install) => install.id === failedInstallId
  )?.workflowHref

  return (
    <div className="flex flex-col gap-6">
      {approval}
      {group.error ? (
        <div className="flex flex-col gap-2">
          <CompositeError error={group.error} />
          {failedInstallHref ? (
            <Link href={failedInstallHref}>View install workflow</Link>
          ) : null}
        </div>
      ) : null}
      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-baseline gap-x-4 gap-y-2">
            <Status status={group.status} />
            {group.doNotDeploy ? (
              <Badge size="sm" theme="neutral">
                Do not deploy
              </Badge>
            ) : null}
            {commit ? <CommitRange commit={commit} /> : null}
            {group.maxParallel != null ? (
              <Text variant="subtext" theme="neutral">
                Up to {group.maxParallel} at a time
              </Text>
            ) : null}
          </div>
          <GroupApproveAllButton
            groupName={group.name}
            installs={group.installs}
          />
        </div>
        <GroupLabels match={group.match} />
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <InstallStatusCounts installs={counted} />
        </div>
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          <SearchInput
            labelClassName="flex-1"
            className="w-full md:min-w-0"
            placeholder="Search installs"
            value={search}
            onChange={setSearch}
          />
          <RadioFilterDropdown
            id={`install-group-${group.id}-status`}
            label="Status"
            options={statuses.map((value) => ({
              value,
              label: humanize(value),
            }))}
            selected={status}
            onChange={setStatus}
          />
        </div>
        {installs.length ? (
          <ul className="flex flex-col gap-2">
            {installs.map((install) => (
              <li key={install.id}>
                <InstallRolloutCard
                  install={install}
                  commit={commit}
                  layout={installLayout}
                  onSelect={onSelectInstall}
                />
              </li>
            ))}
          </ul>
        ) : (
          <Text variant="subtext" theme="neutral">
            No installs match.
          </Text>
        )}
      </div>
    </div>
  )
}
