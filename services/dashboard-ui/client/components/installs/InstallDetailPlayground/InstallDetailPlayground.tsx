import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { BranchTrackingCard } from '@/components/branches/BranchTrackingCard'
import { cn } from '@/utils/classnames'
import { humanize } from '@/utils/string-utils'
import type {
  TPlaygroundInstall,
  TUpdateEvent,
  TLagItem,
  TDriftedObject,
  TStackVersion,
  TRoleEntry,
  TSandboxInfo,
  TComponentEntry,
  TImageEntry,
  TActionEntry,
  TRunbookEntry,
  TUpdateTrigger,
} from './types'

// ─── Header ─────────────────────────────────────────────────────────────────

interface IInstallPlaygroundHeader {
  install: TPlaygroundInstall
}

const BranchChip = ({
  branch,
}: {
  branch: TPlaygroundInstall['appliedBranch']
}) => {
  if (!branch) {
    return (
      <Text variant="subtext" theme="neutral">
        None
      </Text>
    )
  }
  return (
    <span className="flex items-center gap-1.5">
      <Icon
        variant="GitBranchIcon"
        size={13}
        className="text-cool-grey-400 shrink-0"
      />
      <Text variant="subtext" family="mono">
        {branch.name}
      </Text>
      <Badge size="sm" variant="code" theme="neutral">
        {branch.sha.slice(0, 8)}
      </Badge>
    </span>
  )
}

const InstallPlaygroundHeader = ({ install }: IInstallPlaygroundHeader) => {
  const showBranchDelta =
    !install.branchIsCurrent && install.appliedBranch && install.expectedBranch

  return (
    <header className="flex flex-col gap-4 p-4 md:p-6 border-b shrink-0">
      {/* Name row */}
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div className="flex flex-col gap-1.5 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <Text variant="h3" weight="stronger" level={1}>
              {install.name}
            </Text>
            {Object.entries(install.labels).map(([k, v]) => (
              <LabelBadge key={k} size="sm" labelKey={k} labelValue={v} />
            ))}
          </div>

          <div className="flex items-center gap-3 flex-wrap">
            <Text variant="subtext" theme="neutral" family="mono">
              {install.id}
            </Text>
            <Text variant="subtext" theme="info">
              Updated{' '}
              <Time
                time={install.updatedAt}
                format="relative"
                variant="subtext"
              />
            </Text>
          </div>
        </div>

        {/* Right metadata cluster */}
        <div className="flex items-start gap-6 flex-wrap shrink-0">
          {install.isManagedByConfig && (
            <LabeledValue label="Managed by">
              <span className="flex items-center gap-1.5">
                <Icon
                  variant="FileCodeIcon"
                  size={14}
                  className="text-cool-grey-400"
                />
                <Text variant="subtext">Install config</Text>
                {install.configFilePath && (
                  <Badge size="sm" variant="code" theme="neutral">
                    {install.configFilePath}
                  </Badge>
                )}
              </span>
            </LabeledValue>
          )}

          <LabeledValue label="App">
            <Link
              href={`/${install.orgId}/apps/${install.appId}`}
              textVariant="subtext"
            >
              {install.appName}
            </Link>
          </LabeledValue>

          {install.appliedBranch && (
            <LabeledValue label="Applied branch">
              <BranchChip branch={install.appliedBranch} />
            </LabeledValue>
          )}

          {showBranchDelta ? (
            <LabeledValue label="Expected branch">
              <span className="flex items-center gap-1.5">
                <BranchChip branch={install.expectedBranch} />
                <Badge size="sm" theme="warn">
                  Pending
                </Badge>
              </span>
            </LabeledValue>
          ) : null}
        </div>
      </div>
      {install.expectedBranch ? (
        <BranchTrackingCard
          repo={install.expectedBranch.repo}
          branch={install.expectedBranch.repoBranch}
          directory={install.expectedBranch.directory}
          latestRun={{
            status: install.branchIsCurrent ? 'active' : 'pending',
            message: install.expectedBranch.commitMessage,
            author: install.expectedBranch.author,
            sha: install.expectedBranch.sha,
            createdAt: install.expectedBranch.createdAt,
          }}
        />
      ) : null}
    </header>
  )
}

// ─── Updates rail ────────────────────────────────────────────────────────────

const triggerCaption = (trigger?: TUpdateTrigger): string | undefined => {
  if (!trigger) return undefined
  switch (trigger.source) {
    case 'pr':
      return [
        trigger.branch,
        trigger.prNumber && `PR #${trigger.prNumber}`,
        trigger.author && `by ${trigger.author}`,
      ]
        .filter(Boolean)
        .join(' · ')
    case 'push':
      return [
        trigger.branch,
        trigger.sha && trigger.sha.slice(0, 8),
        trigger.author && `by ${trigger.author}`,
      ]
        .filter(Boolean)
        .join(' · ')
    case 'manual':
      return [trigger.author && `by ${trigger.author}`, trigger.branch]
        .filter(Boolean)
        .join(' · ')
  }
}

const UPDATE_TYPE_LABELS: Record<TUpdateEvent['type'], string> = {
  deploy: 'Deploy',
  branch_update: 'Branch',
  config_update: 'Config',
  inputs_update: 'Inputs',
  stack_update: 'Stack',
}

const UpdatesRail = ({ updates }: { updates: TUpdateEvent[] }) => (
  <aside className="flex flex-col w-full max-h-72 shrink-0 border-b overflow-y-auto lg:w-80 lg:max-h-none lg:border-b-0 lg:border-r">
    <div className="flex items-center gap-2 px-4 py-3 border-b">
      <Icon
        variant="ClockCounterClockwiseIcon"
        size={14}
        className="text-cool-grey-400"
      />
      <Text variant="body" weight="strong">
        Updates
      </Text>
    </div>
    <div className="flex flex-col px-4 py-2">
      {updates.map((update) => {
        const caption = triggerCaption(update.trigger)
        return (
          <TimelineEvent
            key={update.id}
            createdAt={update.createdAt}
            status={update.status}
            title={update.title}
            badge={{
              children: UPDATE_TYPE_LABELS[update.type],
              theme: 'neutral',
              size: 'sm',
            }}
            caption={caption ?? update.details}
          />
        )
      })}
    </div>
  </aside>
)

// ─── Overview tab ─────────────────────────────────────────────────────────────

const LagRow = ({ item }: { item: TLagItem }) => (
  <div className="flex items-center justify-between gap-3 py-1.5 min-w-0">
    <Text variant="subtext" className="truncate">
      {item.name}
    </Text>
    <div className="flex items-center gap-2 shrink-0">
      {item.isCurrent ? (
        <Status status="active" variant="badge">
          Current
        </Status>
      ) : (
        <>
          <Badge size="sm" variant="code" theme="neutral">
            {item.appliedVersion.slice(0, 8)}
          </Badge>
          <Icon
            variant="ArrowRightIcon"
            size={12}
            className="text-cool-grey-400"
          />
          <Badge size="sm" variant="code" theme="warn">
            {item.expectedVersion.slice(0, 8)}
          </Badge>
        </>
      )}
    </div>
  </div>
)

const DriftRow = ({ obj }: { obj: TDriftedObject }) => (
  <div className="flex items-center justify-between gap-3 py-1.5">
    <div className="flex items-center gap-2">
      <Status status="warn" isWithoutText variant="timeline" iconSize={14} />
      <Text variant="subtext">
        {obj.targetType === 'sandbox'
          ? 'Sandbox'
          : (obj.componentName ?? 'Component')}
      </Text>
    </div>
    <Badge size="sm" theme="warn">
      Drift detected
    </Badge>
  </div>
)

interface IOverviewTab {
  install: TPlaygroundInstall
}

const OverviewTab = ({ install }: IOverviewTab) => {
  const { configLag, driftedObjects } = install
  const laggedComponents = configLag.components.filter((c) => !c.isCurrent)
  const laggedImages = configLag.images.filter((i) => !i.isCurrent)
  const hasLag =
    (configLag.stack && !configLag.stack.isCurrent) ||
    (configLag.sandbox && !configLag.sandbox.isCurrent) ||
    laggedComponents.length > 0 ||
    laggedImages.length > 0

  const hasDrift = driftedObjects.length > 0

  return (
    <div className="flex flex-col gap-4 p-4">
      {/* Health row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card className="!p-4 !gap-4">
          <div className="flex items-center gap-2">
            <Icon
              variant="SneakerMoveIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Text variant="body" weight="strong">
              Health
            </Text>
          </div>
          <div className="flex flex-col gap-2">
            <LabeledStatus
              label="Runner"
              statusProps={{ status: install.runnerStatus }}
            />
            <LabeledStatus
              label="Sandbox"
              statusProps={{ status: install.sandboxStatus }}
            />
            <LabeledStatus
              label="Components"
              statusProps={{ status: install.componentStatus }}
            />
          </div>
        </Card>

        {/* Branch / config comparison */}
        <Card className="!p-4 !gap-4">
          <div className="flex items-center gap-2">
            <Icon
              variant="GitBranchIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Text variant="body" weight="strong">
              Config
            </Text>
          </div>
          <div className="flex flex-col gap-2">
            {install.appliedBranch && (
              <LabeledValue label="Applied">
                <BranchChip branch={install.appliedBranch} />
              </LabeledValue>
            )}
            {install.expectedBranch && (
              <LabeledValue label="Expected">
                <span className="flex items-center gap-1.5">
                  <BranchChip branch={install.expectedBranch} />
                  {!install.branchIsCurrent && (
                    <Badge size="sm" theme="warn">
                      Pending
                    </Badge>
                  )}
                  {install.branchIsCurrent && (
                    <Badge size="sm" theme="success">
                      Current
                    </Badge>
                  )}
                </span>
              </LabeledValue>
            )}
          </div>
        </Card>

        {/* Drift — explicit separate card, not conflated with config lag */}
        <Card className="!p-4 !gap-4">
          <div className="flex items-center gap-2">
            <Icon
              variant="FileDashedIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Text variant="body" weight="strong">
              Infrastructure drift
            </Text>
          </div>
          {hasDrift ? (
            <div className="flex flex-col divide-y">
              {driftedObjects.map((obj) => (
                <DriftRow key={obj.id} obj={obj} />
              ))}
            </div>
          ) : (
            <div className="flex items-center gap-1.5">
              <Status
                status="active"
                isWithoutText
                variant="timeline"
                iconSize={14}
              />
              <Text variant="subtext" theme="neutral">
                No drift detected
              </Text>
            </div>
          )}
        </Card>
      </div>

      {/* Config lag summary — only shown when config is current (branch matches) but resources are lagging */}
      {install.branchIsCurrent && (
        <Card className="!p-4 !gap-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Icon
                variant="ArrowsClockwiseIcon"
                size={14}
                className="text-cool-grey-400"
              />
              <Text variant="body" weight="strong">
                Config application status
              </Text>
            </div>
            {hasLag ? (
              <Badge size="sm" theme="warn">
                Lagging
              </Badge>
            ) : (
              <Badge size="sm" theme="success">
                All current
              </Badge>
            )}
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-1">
            {/* Stack */}
            {configLag.stack && (
              <div className="flex flex-col">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Stack
                </Text>
                <LagRow item={configLag.stack} />
              </div>
            )}

            {/* Sandbox */}
            {configLag.sandbox && (
              <div className="flex flex-col">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Sandbox
                </Text>
                <LagRow item={configLag.sandbox} />
              </div>
            )}

            {/* Components */}
            {configLag.components.length > 0 && (
              <div className="flex flex-col sm:col-span-2">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Components
                </Text>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8">
                  {configLag.components.map((c) => (
                    <LagRow key={c.name} item={c} />
                  ))}
                </div>
              </div>
            )}

            {/* Images */}
            {configLag.images.length > 0 && (
              <div className="flex flex-col sm:col-span-2">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Images
                </Text>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8">
                  {configLag.images.map((img) => (
                    <LagRow key={img.name} item={img} />
                  ))}
                </div>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* When branch has moved, show the delta separately */}
      {!install.branchIsCurrent && (
        <Card className="!p-4 !gap-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Icon
                variant="ArrowsClockwiseIcon"
                size={14}
                className="text-cool-grey-400"
              />
              <Text variant="body" weight="strong">
                Pending resource changes
              </Text>
            </div>
            <Badge size="sm" theme="warn">
              Branch changed
            </Badge>
          </div>
          <Text variant="subtext" theme="neutral">
            Resources will update once the expected branch is applied. Changes
            below reflect what will be deployed.
          </Text>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-1">
            {configLag.stack && !configLag.stack.isCurrent && (
              <div className="flex flex-col">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Stack
                </Text>
                <LagRow item={configLag.stack} />
              </div>
            )}
            {configLag.sandbox && !configLag.sandbox.isCurrent && (
              <div className="flex flex-col">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Sandbox
                </Text>
                <LagRow item={configLag.sandbox} />
              </div>
            )}
            {laggedComponents.length > 0 && (
              <div className="flex flex-col sm:col-span-2">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Components
                </Text>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8">
                  {laggedComponents.map((c) => (
                    <LagRow key={c.name} item={c} />
                  ))}
                </div>
              </div>
            )}
            {laggedImages.length > 0 && (
              <div className="flex flex-col sm:col-span-2">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Images
                </Text>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8">
                  {laggedImages.map((img) => (
                    <LagRow key={img.name} item={img} />
                  ))}
                </div>
              </div>
            )}
          </div>
        </Card>
      )}
    </div>
  )
}

// ─── Resources sub-tabs ───────────────────────────────────────────────────────

const StackTab = ({ versions }: { versions: TStackVersion[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {versions.map((v) => (
      <Card key={v.id} className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Icon
              variant="StackIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Badge size="sm" variant="code" theme="neutral">
              {v.version}
            </Badge>
            <Badge size="sm" theme="neutral">
              {humanize(v.planType)}
            </Badge>
          </div>
          <div className="flex items-center gap-3">
            <Status status={v.status} variant="badge" />
            <Time
              time={v.createdAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          </div>
        </div>
      </Card>
    ))}
  </div>
)

const RolesTab = ({ roles }: { roles: TRoleEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {roles.map((role) => (
      <Card key={role.id} className="!p-4 !gap-0">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="FileLockIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" family="mono" className="truncate">
              {role.name}
            </Text>
            <Badge size="sm" theme="neutral">
              {humanize(role.type)}
            </Badge>
          </div>
          <Status status={role.status} variant="badge" />
        </div>
      </Card>
    ))}
  </div>
)

const SandboxTab = ({ sandbox }: { sandbox?: TSandboxInfo }) => {
  if (!sandbox) {
    return (
      <div className="p-4">
        <Text variant="subtext" theme="neutral">
          No sandbox configured.
        </Text>
      </div>
    )
  }
  return (
    <div className="p-4">
      <Card className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Icon
              variant="ShippingContainerIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Text variant="body" weight="strong">
              Sandbox
            </Text>
          </div>
          <Status status={sandbox.status} variant="badge" />
        </div>
        <div className="flex flex-wrap gap-x-8 gap-y-3">
          <LabeledValue label="Run type">
            <Badge size="sm" theme="neutral">
              {humanize(sandbox.runType)}
            </Badge>
          </LabeledValue>
          <LabeledValue label="Last run">
            <Time
              time={sandbox.lastRunAt}
              format="relative"
              variant="subtext"
            />
          </LabeledValue>
          {sandbox.workspaceUrl && (
            <LabeledValue label="Workspace">
              <Link
                href={sandbox.workspaceUrl}
                isExternal
                textVariant="subtext"
              >
                View workspace
              </Link>
            </LabeledValue>
          )}
        </div>
      </Card>
    </div>
  )
}

const ComponentsTab = ({ components }: { components: TComponentEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {components.map((cmp) => (
      <Card key={cmp.id} className="!p-4 !gap-0">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="CardsIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" className="truncate">
              {cmp.name}
            </Text>
            <Badge size="sm" theme="neutral">
              {humanize(cmp.type)}
            </Badge>
            {cmp.sha && (
              <Badge size="sm" variant="code" theme="neutral">
                {cmp.sha.slice(0, 8)}
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <Status status={cmp.status} variant="badge" />
            <Time
              time={cmp.deployedAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          </div>
        </div>
      </Card>
    ))}
  </div>
)

const ImagesTab = ({ images }: { images: TImageEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {images.map((img) => (
      <Card key={img.id} className="!p-4 !gap-0">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="PackageIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" family="mono" className="truncate">
              {img.repository}
            </Text>
            <Badge size="sm" variant="code" theme="neutral">
              {img.tag}
            </Badge>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <Status status={img.status} variant="badge" />
            <Time
              time={img.builtAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          </div>
        </div>
      </Card>
    ))}
  </div>
)

interface IResourcesTab {
  resources: TPlaygroundInstall['resources']
}

const ResourcesTab = ({ resources }: IResourcesTab) => (
  <Tabs
    tabs={{
      stack: <StackTab versions={resources.stackVersions} />,
      roles: <RolesTab roles={resources.roles} />,
      sandbox: <SandboxTab sandbox={resources.sandbox} />,
      components: <ComponentsTab components={resources.components} />,
      images: <ImagesTab images={resources.images} />,
    }}
    tabLabels={{
      stack: 'Stack',
      roles: 'Roles',
      sandbox: 'Sandbox',
      components: 'Components',
      images: 'Images',
    }}
    className="gap-0"
    tabControlsClassName="px-4"
  />
)

// ─── Operations sub-tabs ──────────────────────────────────────────────────────

const ActionsTab = ({ actions }: { actions: TActionEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {actions.map((action) => (
      <Card key={action.id} className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="TerminalWindowIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" weight="strong" className="truncate">
              {action.name}
            </Text>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            {action.lastRunStatus && (
              <Status status={action.lastRunStatus} variant="badge" />
            )}
            {action.lastRunAt && (
              <Time
                time={action.lastRunAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            )}
          </div>
        </div>
        <Text variant="subtext" theme="neutral">
          {action.description}
        </Text>
      </Card>
    ))}
  </div>
)

const RunbooksTab = ({ runbooks }: { runbooks: TRunbookEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {runbooks.map((rb) => (
      <Card key={rb.id} className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="BookIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" weight="strong" className="truncate">
              {rb.name}
            </Text>
            <Badge size="sm" theme="neutral">
              {rb.stepCount} {rb.stepCount === 1 ? 'step' : 'steps'}
            </Badge>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            {rb.lastRunStatus && (
              <Status status={rb.lastRunStatus} variant="badge" />
            )}
            {rb.lastRunAt && (
              <Time
                time={rb.lastRunAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            )}
          </div>
        </div>
        <Text variant="subtext" theme="neutral">
          {rb.description}
        </Text>
      </Card>
    ))}
  </div>
)

interface IOperationsTab {
  operations: TPlaygroundInstall['operations']
}

const OperationsTab = ({ operations }: IOperationsTab) => (
  <Tabs
    tabs={{
      actions: <ActionsTab actions={operations.actions} />,
      runbooks: <RunbooksTab runbooks={operations.runbooks} />,
    }}
    tabLabels={{ actions: 'Actions', runbooks: 'Runbooks' }}
    className="gap-0"
    tabControlsClassName="px-4"
  />
)

// ─── Main component ───────────────────────────────────────────────────────────

export interface IInstallDetailPlayground {
  install: TPlaygroundInstall
  className?: string
}

export const InstallDetailPlayground = ({
  install,
  className,
}: IInstallDetailPlayground) => {
  const [activeTab, setActiveTab] = useState<
    'overview' | 'resources' | 'operations'
  >('overview')

  const TAB_LABELS: Record<typeof activeTab, string> = {
    overview: 'Overview',
    resources: 'Resources',
    operations: 'Operations',
  }

  return (
    <div
      className={cn(
        'flex flex-col h-full min-h-0 bg-background border rounded-md overflow-hidden',
        className
      )}
    >
      <InstallPlaygroundHeader install={install} />

      {/* Body: Updates rail (left) + tab content (right) */}
      <div className="flex flex-1 min-h-0 overflow-hidden flex-col lg:flex-row">
        <UpdatesRail updates={install.updates} />

        {/* Main panel */}
        <div className="flex flex-col flex-1 min-w-0 overflow-y-auto">
          {/* Top tab bar */}
          <div
            aria-label="Install sections"
            className="flex items-center gap-6 border-b px-4 shrink-0"
            role="tablist"
          >
            {(Object.keys(TAB_LABELS) as (typeof activeTab)[]).map((key) => (
              <Button
                key={key}
                id={`install-tab-${key}`}
                type="button"
                role="tab"
                aria-controls={`install-tabpanel-${key}`}
                aria-selected={activeTab === key}
                onClick={() => setActiveTab(key)}
                isActive={activeTab === key}
                variant="tab"
              >
                {TAB_LABELS[key]}
              </Button>
            ))}
          </div>

          {/* Tab content */}
          <div
            aria-labelledby={`install-tab-${activeTab}`}
            className="flex-1 min-h-0"
            id={`install-tabpanel-${activeTab}`}
            role="tabpanel"
          >
            {activeTab === 'overview' && <OverviewTab install={install} />}
            {activeTab === 'resources' && (
              <ResourcesTab resources={install.resources} />
            )}
            {activeTab === 'operations' && (
              <OperationsTab operations={install.operations} />
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
