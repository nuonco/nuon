import { useMemo, useState } from 'react'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

export type TAppInstallListStatus = {
  status: string
  detail?: string
}

export type TAppInstallListCommit = {
  sha: string
  message?: string
  appliedAt?: string
  href?: string
  isExternal?: boolean
}

export type TAppInstallListItem = {
  id: string
  name: string
  href: string
  group?: string
  lastAppliedCommit?: TAppInstallListCommit
  deployments: TAppInstallListStatus
  resources: TAppInstallListStatus
  healthChecks: TAppInstallListStatus
}

export interface IAppInstallsList {
  installs: TAppInstallListItem[]
  isLoading?: boolean
  initialSearch?: string
}

const InstallStatus = ({
  label,
  value,
}: {
  label: string
  value: TAppInstallListStatus
}) => (
  <span title={`${label}: ${value.detail ?? value.status}`}>
    <Status status={value.status} variant="badge">
      {label}
    </Status>
  </span>
)

export const AppInstallsList = ({
  installs,
  isLoading = false,
  initialSearch = '',
}: IAppInstallsList) => {
  const [search, setSearch] = useState(initialSearch)
  const filteredInstalls = useMemo(() => {
    const query = search.trim().toLowerCase()
    if (!query) return installs

    return installs.filter((install) =>
      [
        install.name,
        install.id,
        install.group,
        install.lastAppliedCommit?.sha,
        install.lastAppliedCommit?.message,
      ]
        .filter(Boolean)
        .some((value) => value?.toLowerCase().includes(query))
    )
  }, [installs, search])

  return (
    <div className="flex flex-col gap-4">
      <SearchInput
        labelClassName="w-full md:w-fit"
        className="w-full md:w-fit"
        placeholder="Search installs…"
        value={search}
        onChange={setSearch}
        onClear={() => setSearch('')}
      />

      <div className="overflow-hidden rounded-lg border">
        <div className="hidden md:grid grid-cols-12 gap-4 px-4 py-2 bg-cool-grey-100 dark:bg-dark-grey-700">
          <Text variant="label" theme="neutral" className="col-span-3">
            Install name
          </Text>
          <Text variant="label" theme="neutral" className="col-span-2">
            Group
          </Text>
          <Text variant="label" theme="neutral" className="col-span-3">
            Last applied commit
          </Text>
          <Text variant="label" theme="neutral" className="col-span-4">
            Statuses
          </Text>
        </div>

        {isLoading ? (
          Array.from({ length: 4 }).map((_, index) => (
            <div
              key={index}
              className="grid grid-cols-12 gap-4 px-4 py-3 border-t"
            >
              <Skeleton className="col-span-3" height="20px" />
              <Skeleton className="col-span-2" height="20px" />
              <Skeleton className="col-span-3" height="20px" />
              <Skeleton className="col-span-4" height="20px" />
            </div>
          ))
        ) : filteredInstalls.length > 0 ? (
          filteredInstalls.map((install) => (
            <div
              key={install.id}
              className="grid grid-cols-1 md:grid-cols-12 gap-2 md:gap-4 px-4 py-3 border-t first:border-t-0 md:first:border-t"
            >
              <div className="md:col-span-3 min-w-0 flex flex-col">
                <Link href={install.href} className="font-mono truncate">
                  {install.name}
                </Link>
                <Text
                  variant="label"
                  theme="neutral"
                  family="mono"
                  className="truncate"
                >
                  {install.id}
                </Text>
              </div>
              <Text
                variant="subtext"
                theme="neutral"
                family="mono"
                className="md:col-span-2 truncate"
              >
                {install.group || '—'}
              </Text>
              <div className="md:col-span-3 min-w-0">
                {install.lastAppliedCommit ? (
                  <BranchRunCommit
                    displayVariant="inline"
                    showStatus={false}
                    sha={install.lastAppliedCommit.sha}
                    message={install.lastAppliedCommit.message}
                    createdAt={install.lastAppliedCommit.appliedAt}
                    href={install.lastAppliedCommit.href}
                    isExternal={install.lastAppliedCommit.isExternal}
                  />
                ) : (
                  <Text variant="subtext" theme="neutral">
                    No commit applied
                  </Text>
                )}
              </div>
              <div className="md:col-span-4 flex flex-wrap items-center gap-2">
                <InstallStatus
                  label="Deployments"
                  value={install.deployments}
                />
                <InstallStatus label="Resources" value={install.resources} />
                <InstallStatus
                  label="Health checks"
                  value={install.healthChecks}
                />
              </div>
            </div>
          ))
        ) : (
          <EmptyState
            variant="diagram"
            emptyTitle={search ? 'No installs found' : 'No installs yet'}
            emptyMessage={
              search
                ? 'Try a different install name, group, or ID.'
                : 'Installs will appear here after they are created.'
            }
          />
        )}
      </div>
    </div>
  )
}
