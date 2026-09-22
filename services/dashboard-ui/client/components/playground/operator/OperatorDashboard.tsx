import { useMemo, useState, type ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CompositeError } from '@/components/common/CompositeError'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { StatusWithDescription } from '@/components/common/StatusWithDescription'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { PageContent } from '@/components/layout/PageContent'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PendingApprovalsComponent } from '@/components/orgs/PendingApprovals'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TCloudPlatform } from '@/types'
import {
  operatorActiveWorkflows,
  operatorApprovals,
  operatorInstalls,
  OPERATOR_ORG_ID,
  type TOperatorInstall,
} from './fixtures'

type TOperatorFilter = 'all' | 'attention' | 'behind' | 'current'

const ATTENTION_STATUSES = [
  'failed',
  'error',
  'degraded',
  'unhealthy',
  'drifted',
  'unknown',
]

const isBehind = (install: TOperatorInstall) =>
  Boolean(
    install.app_config_version &&
      install.app_config_latest_version &&
      install.app_config_latest_version > install.app_config_version
  )

const needsAttention = (install: TOperatorInstall) =>
  [
    install.deployments_status,
    install.resources_status,
    install.health_status,
  ].some((status) => ATTENTION_STATUSES.includes(status ?? ''))

const regionOf = (install: TOperatorInstall) =>
  install.aws_account?.region ||
  install.gcp_account?.region ||
  install.azure_account?.location ||
  '—'

const StatusCell = ({
  install,
  status,
  detail,
}: {
  install: TOperatorInstall
  status?: string
  detail?: string
}) => {
  const { addPanel } = useSurfaces()

  if (!status) {
    return (
      <Text variant="subtext" theme="neutral">
        —
      </Text>
    )
  }

  const errors = install.errors ?? []
  const canOpenErrors = ATTENTION_STATUSES.includes(status) && errors.length > 0

  if (!canOpenErrors) {
    return (
      <StatusWithDescription
        statusProps={{ variant: 'default', status }}
        tooltipProps={{ position: 'top', tipContent: detail }}
      />
    )
  }

  const isFatal = errors.some(
    (e) => e.severity === 'fatal' || e.severity === 'error'
  )

  const primaryAction = errors.find((e) => e.action)?.action

  const openErrors = () =>
    addPanel(
      <Panel
        heading={`${install.name} · errors`}
        size="half"
        footer={
          primaryAction ? (
            <Button variant="primary" href={primaryAction.href}>
              {primaryAction.label}
            </Button>
          ) : undefined
        }
      >
        <div className="flex flex-col gap-6">
          {errors.map((error, i) => (
            <div className="flex flex-col gap-2" key={`${error.type}-${i}`}>
              <CompositeError error={error} />
              {error.action ? (
                <Link href={error.action.href} variant="inline">
                  {error.action.label}
                </Link>
              ) : null}
            </div>
          ))}
        </div>
      </Panel>
    )

  return (
    <Button
      variant="ghost"
      className="!h-auto !py-1 !px-1.5 !rounded-md"
      tooltipProps={{ position: 'top', tipContent: detail }}
      onClick={openErrors}
    >
      <span className="flex items-center gap-1.5 whitespace-nowrap">
        <Status variant="default" status={status} />
        <Badge size="xs" theme={isFatal ? 'error' : 'warn'}>
          {errors.length}
        </Badge>
      </span>
    </Button>
  )
}

const VersionCell = ({ install }: { install: TOperatorInstall }) => {
  const current = install.app_config_version
  const latest = install.app_config_latest_version

  if (!current) {
    return (
      <Text variant="subtext" theme="neutral">
        —
      </Text>
    )
  }

  if (!isBehind(install)) {
    return (
      <Badge size="sm" variant="code" theme="default">
        v{current}
      </Badge>
    )
  }

  return (
    <span className="flex items-center gap-1 whitespace-nowrap">
      <Badge size="sm" variant="code" theme="warn">
        v{current}
      </Badge>
      <Icon variant="ArrowRightIcon" size={11} theme="neutral" />
      <Badge size="sm" variant="code" theme="neutral">
        v{latest}
      </Badge>
    </span>
  )
}

const ActionsCell = ({ install }: { install: TOperatorInstall }) => {
  const href = `/${OPERATOR_ORG_ID}/installs/${install.id}`

  return (
    <Dropdown
      alignment="right"
      buttonText=""
      buttonClassName="!p-1"
      icon={<Icon variant="DotsThreeVerticalIcon" />}
      id={install.id ?? ''}
      variant="ghost"
    >
      <Menu>
        <Button href={href}>View details</Button>
        <Button href={`${href}/history`}>Open latest run</Button>
        <hr />
        <Text variant="label" theme="neutral">
          Deploy
        </Text>
        {isBehind(install) ? (
          <Button isMenuButton>
            Update to v{install.app_config_latest_version}
            <Icon variant="ArrowUpIcon" />
          </Button>
        ) : null}
        {install.deployments_status === 'pending-approval' ? (
          <Button isMenuButton>
            Review approval
            <Icon variant="CheckCircleIcon" />
          </Button>
        ) : null}
        <Button isMenuButton>
          Redeploy current version
          <Icon variant="ArrowClockwiseIcon" />
        </Button>
        <Button isMenuButton>
          Reprovision sandbox
          <Icon variant="BoxArrowUpIcon" />
        </Button>
        <hr />
        <Text variant="label" theme="neutral">
          Inspect
        </Text>
        {install.resources_status === 'drifted' ? (
          <Button isMenuButton>
            Review drift
            <Icon variant="WarningIcon" />
          </Button>
        ) : null}
        <Button href={`${href}/inputs`} isMenuButton>
          Current inputs
          <Icon variant="ListChecksIcon" />
        </Button>
        <Button href={`${href}/state`} isMenuButton>
          View state
          <Icon variant="CodeBlockIcon" />
        </Button>
        <hr />
        <Button isMenuButton variant="danger">
          Deprovision
        </Button>
      </Menu>
    </Dropdown>
  )
}

type TOperatorRow = {
  actions: ReactNode
  cloud: ReactNode
  deployments: ReactNode
  health: ReactNode
  install: ReactNode
  name: string
  resources: ReactNode
  version: ReactNode
}

const toRows = (installs: TOperatorInstall[]): TOperatorRow[] =>
  installs.map((install) => ({
    name: install.name ?? '',
    install: (
      <span className="flex flex-col gap-0.5 min-w-0">
        <Link
          href={`/${OPERATOR_ORG_ID}/installs/${install.id}`}
          variant="inline"
        >
          {install.name}
        </Link>
        <Text variant="caption" theme="neutral" className="whitespace-nowrap">
          {install.app?.name}
          {install.app_branch?.name ? ` · ${install.app_branch.name}` : ''}
        </Text>
      </span>
    ),
    version: <VersionCell install={install} />,
    deployments: (
      <StatusCell
        install={install}
        status={install.deployments_status}
        detail={install.deployments_detail}
      />
    ),
    resources: (
      <StatusCell
        install={install}
        status={install.resources_status}
        detail={install.resources_detail}
      />
    ),
    health: (
      <StatusCell
        install={install}
        status={install.health_status}
        detail={install.health_detail}
      />
    ),
    cloud: (
      <span className="flex items-center gap-2 whitespace-nowrap">
        <CloudPlatform
          platform={(install.cloud_platform as TCloudPlatform) || 'unknown'}
          variant="subtext"
          colorVariant="color"
          displayVariant="icon-only"
          iconSize="18"
        />
        <Text variant="caption" family="mono" theme="neutral">
          {regionOf(install)}
        </Text>
      </span>
    ),
    actions: <ActionsCell install={install} />,
  }))

const columns: ColumnDef<TOperatorRow>[] = [
  {
    accessorKey: 'name',
    header: 'Install',
    cell: (info) => info.row.original.install,
    enableSorting: true,
  },
  {
    accessorKey: 'version',
    header: 'Version',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'deployments',
    header: 'Deployments',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'resources',
    header: 'Resources',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'health',
    header: 'Health',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'cloud',
    header: 'Cloud',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'actions',
    header: '',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
]

export const OperatorDashboard = ({
  installs = operatorInstalls,
  initialFilter,
}: {
  installs?: TOperatorInstall[]
  initialFilter?: TOperatorFilter
}) => {
  const [filter, setFilter] = useState<TOperatorFilter>(
    () => initialFilter ?? (installs.some(needsAttention) ? 'attention' : 'all')
  )

  const counts = useMemo(
    () => ({
      all: installs.length,
      attention: installs.filter(needsAttention).length,
      behind: installs.filter(isBehind).length,
      current: installs.filter((i) => !isBehind(i) && !needsAttention(i)).length,
    }),
    [installs]
  )

  const filtered = useMemo(() => {
    switch (filter) {
      case 'attention':
        return installs.filter(needsAttention)
      case 'behind':
        return installs.filter(isBehind)
      case 'current':
        return installs.filter((i) => !isBehind(i) && !needsAttention(i))
      default:
        return installs
    }
  }, [filter, installs])

  return (
    <>
      <SectionHeader
        variant="page"
        title="Installs"
        description="Every customer install, the version it runs, and what needs attention."
      />
      <PageContent>
        <PageSection>
          <PendingApprovalsComponent
            orgId={OPERATOR_ORG_ID}
            approvals={operatorApprovals}
            activeWorkflows={operatorActiveWorkflows}
          />
          <Table<TOperatorRow>
            columns={columns}
            data={toRows(filtered)}
            enableSearch
            searchPlaceholder="Search installs"
            filterActions={
              <ToggleButton<TOperatorFilter>
                size="lg"
                label="Install filter"
                value={filter}
                onChange={setFilter}
                options={[
                  { value: 'all', label: `All ${counts.all}` },
                  {
                    value: 'attention',
                    label: `Needs attention ${counts.attention}`,
                  },
                  { value: 'behind', label: `Out of date ${counts.behind}` },
                  { value: 'current', label: `Healthy ${counts.current}` },
                ]}
              />
            }
            emptyStateProps={{
              emptyTitle: 'No installs match this filter',
              emptyMessage: 'Clear the filter to see every install in the org.',
            }}
          />
        </PageSection>
      </PageContent>
    </>
  )
}
