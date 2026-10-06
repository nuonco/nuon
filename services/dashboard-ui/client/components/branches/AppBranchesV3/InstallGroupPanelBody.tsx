import { useState } from 'react'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { humanize } from '@/utils/string-utils'
import { CommitSha } from './CommitSha'
import { GroupLabels } from './GroupLabels'
import type { TLatestRollout, TRolloutInstallGroup } from './fixtures'
import { InstallRolloutPanel } from './InstallRolloutPanel'
import { InstallStatusCounts } from './InstallStatusCounts'

export const InstallGroupPanelBody = ({
  group,
  commit,
}: {
  group: TRolloutInstallGroup
  commit: TLatestRollout['commit']
}) => {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<string>()
  const statuses = [...new Set(group.installs.map((install) => install.status))]
  const query = search.trim().toLowerCase()
  const installs = group.installs.filter(
    (install) =>
      (!query || install.name.toLowerCase().includes(query)) &&
      (!status || install.status === status)
  )

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
          <Status status={group.status} />
          <CommitSha commit={commit} />
          <Text variant="subtext" theme="neutral">
            Up to {group.max_parallel ?? 1} at a time
          </Text>
        </div>
        <GroupLabels group={group} />
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <InstallStatusCounts group={group} />
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
                <InstallRolloutPanel install={install} />
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
