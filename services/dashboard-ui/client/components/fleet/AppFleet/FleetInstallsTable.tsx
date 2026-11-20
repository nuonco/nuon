import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { useInstallHref } from '@/hooks/use-install-path'
import type { TAppFleetInstall } from '@/types'
import { NO_SIGNAL, ROLLOUT_OPTIONS } from './fleet-utils'

const Missing = () => <Icon variant="MinusIcon" />

export const RolloutBadge = ({ rollout }: { rollout?: string }) => {
  if (rollout !== 'applied' && rollout !== 'behind') return <Missing />
  return (
    <Badge size="sm" theme={rollout === 'applied' ? 'success' : 'warn'}>
      {ROLLOUT_OPTIONS.find((option) => option.value === rollout)?.label}
    </Badge>
  )
}

const InstallName = ({
  install,
  orgId,
  appId,
}: {
  install: TAppFleetInstall
  orgId?: string
  appId?: string
}) => {
  const installHref = useInstallHref()
  return (
    <span>
      <Text variant="body">
        <Link
          href={installHref({ orgId, appId, installId: install.id })}
          variant="inline"
        >
          {install.name}
        </Link>
      </Text>
      <ID>{install.id}</ID>
    </span>
  )
}

const buildColumns = ({
  orgId,
  appId,
  labelColors,
  showLabels,
}: {
  orgId?: string
  appId?: string
  labelColors?: Record<string, string>
  showLabels: boolean
}): ColumnDef<TAppFleetInstall>[] => [
  {
    accessorKey: 'name',
    header: 'Install name',
    meta: { className: 'w-1/4' },
    cell: (info) => (
      <InstallName install={info.row.original} orgId={orgId} appId={appId} />
    ),
  },
  {
    id: 'rollout',
    header: 'Latest run',
    enableSorting: false,
    meta: { className: 'w-32' },
    cell: (info) => <RolloutBadge rollout={info.row.original.rollout} />,
  },
  {
    id: 'health',
    header: 'Health',
    enableSorting: false,
    cell: (info) => {
      const health = info.row.original.health
      return !health || health === NO_SIGNAL ? (
        <Missing />
      ) : (
        <Status status={health} />
      )
    },
  },
  {
    id: 'runner',
    header: 'Runner',
    enableSorting: false,
    cell: (info) => (
      <Status status={info.row.original.runner_status || 'unknown'} />
    ),
  },
  ...(showLabels
    ? [
        {
          id: 'labels',
          header: 'Labels',
          enableSorting: false,
          cell: (info) => {
            const labels = info.row.original.labels ?? {}
            const keys = Object.keys(labels).sort()
            if (!keys.length) return <Missing />
            return (
              <span className="flex flex-wrap gap-1">
                {keys.map((key) => (
                  <LabelBadge
                    key={key}
                    size="sm"
                    labelKey={key}
                    labelValue={labels[key]}
                    customColor={labelColors?.[key]}
                  />
                ))}
              </span>
            )
          },
        } satisfies ColumnDef<TAppFleetInstall>,
      ]
    : []),
]

export interface IFleetInstallsTable {
  installs: TAppFleetInstall[]
  orgId?: string
  appId?: string
  labelColors?: Record<string, string>
  showLabels?: boolean
  isLoading?: boolean
}

export const FleetInstallsTable = ({
  installs,
  orgId,
  appId,
  labelColors,
  showLabels = true,
  isLoading,
}: IFleetInstallsTable) => (
  <Table<TAppFleetInstall>
    columns={buildColumns({ orgId, appId, labelColors, showLabels })}
    data={installs}
    enableSearch={false}
    enableSorting={false}
    isLoading={isLoading}
  />
)
