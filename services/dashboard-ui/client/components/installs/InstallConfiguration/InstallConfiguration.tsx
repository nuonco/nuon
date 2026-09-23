import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { CodeBlock } from '@/components/diffs/CodeBlock'
import { ComponentOverridesList } from '@/components/install-overrides/ComponentOverridesList'
import { InputValue } from '@/components/installs/management/InputValue'
import type { TAppBranchConfig, TAppBranchRun, TAppInput } from '@/types'
import { getInputDisplayName } from '@/utils/install-utils'

type TInputGroup = {
  id?: string
  name?: string
  display_name?: string
  description?: string
  app_inputs?: TAppInput[]
}

const runCommit = (
  run: TAppBranchRun | undefined,
  href: string | undefined
) => {
  const commit = run?.vcs_connection_commit

  return run ? (
    <BranchRunCommit
      status={run.status}
      href={href}
      message={commit?.message?.split('\n')[0]}
      author={commit?.author_name}
      avatarUrl={commit?.author_avatar_url}
      sha={commit?.sha ?? run.head_sha}
      createdAt={run.created_at}
    />
  ) : null
}

const RunCard = ({
  emptyMessage,
  href,
  label,
  run,
}: {
  emptyMessage: string
  href?: string
  label: string
  run?: TAppBranchRun
}) => (
  <Card className="!p-4 !gap-3">
    <Text variant="body" weight="strong">
      {label}
    </Text>
    {run ? (
      runCommit(run, href)
    ) : (
      <Text variant="subtext" theme="neutral">
        {emptyMessage}
      </Text>
    )}
  </Card>
)

export interface IInstallConfigurationAppBranch {
  appliedConfigId?: string
  appliedRun?: TAppBranchRun
  appliedRunHref?: string
  branchConfig?: TAppBranchConfig
  branchHref?: string
  branchName?: string
  history?: ReactNode
  isLoading?: boolean
  latestRun?: TAppBranchRun
  latestRunHref?: string
}

export const InstallConfigurationAppBranch = ({
  appliedConfigId,
  appliedRun,
  appliedRunHref,
  branchConfig,
  branchHref,
  branchName,
  history,
  isLoading,
  latestRun,
  latestRunHref,
}: IInstallConfigurationAppBranch) => {
  if (isLoading && !branchName) {
    return <Skeleton height="180px" width="100%" />
  }

  if (!branchName) {
    return (
      <EmptyState
        variant="history"
        emptyTitle="No app branch connected"
        emptyMessage="Connect this install to an app branch to track its applied configuration."
      />
    )
  }

  const vcs =
    branchConfig?.connected_github_vcs_config ??
    branchConfig?.public_git_vcs_config
  const repoHref = vcs?.repo
    ? (vcs.repo.startsWith('http')
        ? vcs.repo
        : `https://github.com/${vcs.repo}`
      ).replace(/\.git\/?$/, '')
    : undefined
  const isCurrent = !!latestRun?.id && latestRun.id === appliedRun?.id

  return (
    <div className="flex flex-col gap-4">
      <Card className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <Text variant="body" weight="strong">
            Tracking
          </Text>
          <Status status={isCurrent ? 'active' : 'pending'} variant="badge">
            {isCurrent ? 'Current' : 'Update available'}
          </Status>
        </div>
        <div className="flex flex-wrap gap-x-8 gap-y-4">
          <LabeledValue label="App branch">
            {branchHref ? (
              <Link href={branchHref} textVariant="subtext">
                {branchName}
              </Link>
            ) : (
              <Text variant="subtext">{branchName}</Text>
            )}
          </LabeledValue>
          {appliedConfigId ? (
            <LabeledValue label="Applied config">
              <ID>{appliedConfigId}</ID>
            </LabeledValue>
          ) : null}
          {vcs?.repo ? (
            <LabeledValue label="Repository">
              {repoHref ? (
                <Link href={repoHref} isExternal textVariant="subtext">
                  {vcs.repo}
                </Link>
              ) : (
                <Text variant="subtext">{vcs.repo}</Text>
              )}
            </LabeledValue>
          ) : null}
          {vcs?.branch ? (
            <LabeledValue label="Git branch">
              <Text variant="subtext" family="mono">
                {vcs.branch}
              </Text>
            </LabeledValue>
          ) : null}
          {vcs?.directory && vcs.directory !== '.' ? (
            <LabeledValue label="Directory">
              <Text variant="subtext" family="mono">
                {vcs.directory}
              </Text>
            </LabeledValue>
          ) : null}
        </div>
      </Card>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <RunCard
          label="Expected / latest run"
          run={latestRun}
          href={latestRunHref}
          emptyMessage="This branch has not run yet."
        />
        <RunCard
          label="Currently applied"
          run={appliedRun}
          href={appliedRunHref}
          emptyMessage="No branch run has been applied to this install yet."
        />
      </div>

      {history ? (
        <div className="flex flex-col gap-3">
          <Text variant="body" weight="strong">
            Applied config history
          </Text>
          {history}
        </div>
      ) : null}
    </div>
  )
}

export interface IInstallConfigurationInputs {
  action?: ReactNode
  groups: TInputGroup[]
  isLoading?: boolean
  values?: Record<string, string>
}

export const InstallConfigurationInputs = ({
  action,
  groups,
  isLoading,
  values = {},
}: IInstallConfigurationInputs) => {
  if (isLoading) return <Skeleton height="180px" width="100%" />

  if (!groups.length) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No inputs configured"
        emptyMessage="This app config does not define any install inputs."
      />
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {action ? <div className="flex justify-end">{action}</div> : null}
      {groups.map((group) => {
        const inputs = group.app_inputs ?? []

        return (
          <Card key={group.id ?? group.name} className="!p-4 !gap-4">
            <div className="flex flex-col gap-0.5">
              <Text variant="body" weight="strong">
                {group.display_name ?? group.name}
              </Text>
              {group.description ? (
                <Text variant="subtext" theme="neutral">
                  {group.description}
                </Text>
              ) : null}
            </div>
            <PropertyGrid
              align="start"
              columns={[
                { key: 'input', header: 'Input' },
                { key: 'value', header: 'Current value' },
                { key: 'defaultValue', header: 'Default' },
              ]}
              gridTemplate="minmax(150px, 1fr) minmax(150px, 2fr) minmax(120px, 1fr)"
              values={inputs.map((input) => ({
                input: (
                  <span className="flex flex-col">
                    <Text variant="subtext" weight="strong">
                      {input.display_name ?? input.name}
                    </Text>
                    <Text variant="label" family="mono" theme="neutral">
                      {input.name ? getInputDisplayName(input.name) : null}
                    </Text>
                  </span>
                ),
                value: (
                  <InputValue
                    name={input.name}
                    value={input.name ? values[input.name] : undefined}
                  />
                ),
                defaultValue: (
                  <Text variant="label" family="mono" theme="neutral">
                    {input.default ?? '—'}
                  </Text>
                ),
              }))}
            />
          </Card>
        )
      })}
    </div>
  )
}

export interface IInstallConfigurationOverrides {
  action?: ReactNode
  inputs?: TAppInput[]
  isLoading?: boolean
  values?: Record<string, string>
}

export const InstallConfigurationOverrides = ({
  action,
  inputs,
  isLoading,
  values,
}: IInstallConfigurationOverrides) => {
  if (isLoading) return <Skeleton height="180px" width="100%" />

  return (
    <div className="flex flex-col gap-4">
      {action ? <div className="flex justify-end">{action}</div> : null}
      <ComponentOverridesList
        inputs={inputs}
        values={values}
        codeBlockVariant="viewer"
      />
    </div>
  )
}

export interface IInstallConfigurationConfigFile {
  action?: ReactNode
  content?: string
  filename?: string
  history?: ReactNode
  isLoading?: boolean
  isManagedByConfig: boolean
  latestVersionId?: string
  syncedAt?: string
}

export const InstallConfigurationConfigFile = ({
  action,
  content,
  filename,
  history,
  isLoading,
  isManagedByConfig,
  latestVersionId,
  syncedAt,
}: IInstallConfigurationConfigFile) => {
  if (!isManagedByConfig) {
    return (
      <EmptyState
        variant="diagram"
        emptyTitle="No install config file"
        emptyMessage="This install is managed from the dashboard."
      />
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Card className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3 flex-wrap">
          <span className="flex items-center gap-2 min-w-0">
            <Icon
              variant="FileCodeIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" family="mono" className="truncate">
              {filename ?? 'install.toml'}
            </Text>
          </span>
          {latestVersionId ? (
            <Badge size="sm" variant="code" theme="neutral">
              {latestVersionId}
            </Badge>
          ) : null}
        </div>
        {syncedAt ? (
          <LabeledValue label="Last synced">
            <Time time={syncedAt} format="relative" variant="subtext" />
          </LabeledValue>
        ) : null}
        {isLoading ? (
          <Skeleton height="240px" width="100%" />
        ) : content ? (
          <CodeBlock value={content} language="toml" filename={filename} copy />
        ) : (
          <Text variant="subtext" theme="neutral">
            The current install config could not be generated.
          </Text>
        )}
      </Card>

      <div className="flex items-center justify-between gap-3">
        <Text variant="body" weight="strong">
          Config history
        </Text>
        {action}
      </div>
      {history}
    </div>
  )
}

export const ConfigSyncAction = ({
  isPending,
  onSync,
}: {
  isPending?: boolean
  onSync: () => void
}) => (
  <Button variant="secondary" disabled={isPending} onClick={onSync}>
    {isPending ? 'Syncing config' : 'Sync now'}
  </Button>
)
