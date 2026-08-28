import { useSearchParams } from 'react-router'
import { PlanComponent } from '@/components/approvals/Plan'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BundleContentsTable } from '@/components/apps/bundles/BundleContentsTable'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TCustomerManagedBundleArtifact } from '@/types'
import type { TCustomerManagedBundleInfo } from '@/lib'
import { formatBytes } from '@/utils/string-utils'
import type { ColumnDef } from '@tanstack/react-table'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'
import { SnapshotCloudFormationPlan } from './SnapshotCloudFormationPlan'
import { toCustomerManagedWorkflowStep } from './SnapshotRunStepDetails'
import { SnapshotStagedBundleDiff } from './SnapshotStagedBundleDiff'

const historyColumns: ColumnDef<TCustomerManagedBundleInfo>[] = [
  {
    accessorKey: 'bundle_digest',
    header: 'Bundle',
    cell: ({ row }) => <Hash hash={row.original.bundle_digest} />,
  },
  {
    accessorKey: 'activated_at',
    header: 'Activated',
    cell: ({ row }) => (
      <Time time={row.original.activated_at} format="long-datetime" />
    ),
  },
  {
    accessorKey: 'contents',
    header: 'Contents',
    cell: ({ row }) => row.original.contents?.length ?? 0,
  },
  {
    accessorKey: 'total_size',
    header: 'Size',
    cell: ({ row }) =>
      row.original.total_size ? formatBytes(row.original.total_size) : '—',
  },
]

export const CustomerManagedSnapshotBundles = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()
  const active = snapshot?.snapshot.active_bundle
  const staged = snapshot?.snapshot.staged_bundle
  const history = snapshot?.snapshot.bundle_history ?? []
  const planRuns = (snapshot?.snapshot.runs ?? []).filter(
    ({ ref_kind, steps }) =>
      ref_kind === 'bundle-plan' && steps.some(({ plan }) => !!plan)
  )
  const snapshotId = searchParams.get('snapshot')
  const publishedBundleHref =
    install.app_id && active?.release?.id
      ? `/${org.id}/apps/${install.app_id}/releases/${active.release.id}`
      : undefined
  const artifacts: TCustomerManagedBundleArtifact[] = (
    active?.contents ?? []
  ).map((content, index) => ({
    id: `${active?.bundle_digest ?? 'bundle'}-${index}`,
    kind: content.kind.replaceAll('-', '_'),
    logical_name:
      content.kind === 'stack-asset' &&
      (content.name === 'root' || content.detail?.startsWith('compiled:'))
        ? 'Install stack'
        : content.name,
    repository: content.detail,
    digest: content.digest,
    config_digest: content.config_digest,
    size: content.size,
  }))

  const stagedBundleContent = staged ? (
    <div className="flex flex-col gap-6 mt-6">
      <SnapshotStagedBundleDiff candidate={staged} />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Captured deployment plans
        </Text>
        <Text variant="subtext" theme="neutral">
          Plans generated for this staged bundle in the customer environment. A
          plan does not activate the bundle.
        </Text>
      </HeadingGroup>
      {planRuns.length ? (
        <div className="flex flex-col gap-4">
          {planRuns.map((run, runIndex) => {
            const runHref = `/${org.id}/installs/${install.id}/workflows/${run.run_id}${snapshotId ? `?snapshot=${snapshotId}` : ''}`
            return (
              <Card key={run.run_id}>
                <div className="flex flex-wrap items-start justify-between gap-4">
                  <HeadingGroup>
                    <span className="flex items-center gap-2">
                      <Text weight="strong">Bundle deployment plan</Text>
                      <Badge theme="neutral">Staged only</Badge>
                    </span>
                    {run.bundle_digest ? (
                      <Hash hash={run.bundle_digest} length={20} />
                    ) : null}
                    <Time time={run.started_at} format="long-datetime" />
                  </HeadingGroup>
                  <span className="flex items-center gap-3">
                    <Status
                      status={
                        run.status === 'finished' ? 'success' : run.status
                      }
                      variant="badge"
                    />
                    <Link href={runHref} className="font-strong">
                      View run
                      <Icon variant="ArrowRightIcon" size={14} />
                    </Link>
                  </span>
                </div>
                <div className="flex flex-col gap-3">
                  {run.steps
                    .filter(({ plan }) => !!plan)
                    .map((step, stepIndex) => (
                      <Expand
                        id={`captured-bundle-plan-${runIndex}-${stepIndex}`}
                        isOpen
                        className="border rounded-md"
                        headerClassName="px-4 py-3"
                        heading={
                          <div className="flex flex-1 items-center justify-between gap-4 text-left">
                            <Text weight="strong">{step.name}</Text>
                            <Badge variant="code">{step.plan!.kind}</Badge>
                          </div>
                        }
                        key={step.id}
                      >
                        <div className="border-t p-4">
                          {step.plan!.kind === 'cloudformation' ? (
                            <SnapshotCloudFormationPlan
                              plan={step.plan!.content}
                            />
                          ) : (
                            <PlanComponent
                              step={toCustomerManagedWorkflowStep(step)}
                              plan={step.plan!.content}
                              isLoading={false}
                              error={undefined}
                            />
                          )}
                        </div>
                      </Expand>
                    ))}
                </div>
              </Card>
            )
          })}
        </div>
      ) : (
        <Text theme="neutral">
          No deployment plan was included in this support snapshot. Run a plan
          in the customer portal, then capture and upload a new support bundle.
        </Text>
      )}
    </div>
  ) : null

  const activeBundleContent = (
    <div className="flex flex-col gap-6 mt-6">
      {active ? (
        <Card>
          <div className="flex flex-wrap items-start justify-between gap-4">
            <HeadingGroup>
              <Text variant="base" weight="strong">
                Active bundle
              </Text>
              <Hash hash={active.bundle_digest} length={20} />
            </HeadingGroup>
            <span className="flex flex-wrap items-center gap-3">
              {publishedBundleHref ? (
                <Link href={publishedBundleHref} className="font-strong">
                  View bundle diff
                  <Icon variant="ArrowSquareOutIcon" size={14} />
                </Link>
              ) : null}
              <Status status="active" variant="badge" />
            </span>
          </div>
          <BundleContentsTable
            artifacts={artifacts}
            getArtifactHref={(artifact) => {
              const installPath = `/${org.id}/installs/${install.id}`
              if (artifact.kind === 'component') {
                return `${installPath}/components/${artifact.logical_name}`
              }
              if (artifact.kind === 'sandbox') return `${installPath}/sandbox`
              if (artifact.kind === 'stack_asset')
                return `${installPath}/stacks`
              if (
                artifact.kind === 'runner_binary' ||
                artifact.kind === 'runner_image'
              ) {
                return `${installPath}/runner`
              }
              return undefined
            }}
          />
        </Card>
      ) : (
        <Text theme="neutral">
          Active bundle metadata was unavailable when this snapshot was
          captured. Capture and upload a new support bundle to refresh it.
        </Text>
      )}
    </div>
  )

  const historyContent = (
    <div className="flex flex-col gap-6 mt-6">
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Activation history
        </Text>
        <Text variant="subtext" theme="neutral">
          Each row is an immutable bundle activation record captured from the
          customer environment.
        </Text>
      </HeadingGroup>
      <Table
        columns={historyColumns}
        data={history}
        enableSearch={false}
        emptyStateProps={{
          variant: 'table',
          emptyTitle: 'No bundle history captured',
          emptyMessage:
            'Bundle activations will appear after an upgrade and a new support snapshot.',
        }}
      />
    </div>
  )

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Bundle state
        </Text>
        <Text variant="subtext" theme="neutral">
          Review the staged bundle, current active bundle, and activation
          history.
        </Text>
      </HeadingGroup>
      <Tabs
        initActiveTab={staged ? 'staged bundle' : 'active bundle'}
        tabs={{
          ...(stagedBundleContent
            ? { 'staged bundle': stagedBundleContent }
            : {}),
          'active bundle': activeBundleContent,
          history: historyContent,
        }}
      />
    </CustomerManagedSnapshotContent>
  )
}
