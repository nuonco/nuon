import { useMemo, type ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TAppLabelKey } from '@/lib/ctl-api/apps/get-app-labels'
import { LabelColorPicker } from './LabelColorPicker'

const OTHER_LABELS_DESCRIPTION =
  'One-off labels applied to components, actions, runbooks, and installs of this app. Each key gets a color automatically — override any you want to customize.'

interface IDefaultLabelRow {
  key: string
  value: string
  color?: string
}

interface IAppLabels {
  labels: TAppLabelKey[]
  defaultLabels?: Record<string, string>
  isLoading?: boolean
  isPending?: boolean
  resetAction?: ReactNode
  onOverride: (key: string, color: string) => void
  onRemoveOverride: (key: string) => void
}

export const AppLabels = ({
  labels,
  defaultLabels = {},
  isLoading,
  isPending,
  resetAction,
  onOverride,
  onRemoveOverride,
}: IAppLabels) => {
  const defaultRows = useMemo<IDefaultLabelRow[]>(
    () =>
      Object.entries(defaultLabels)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([key, value]) => ({
          key,
          value,
          color: labels.find((label) => label.key === key)?.color,
        })),
    [defaultLabels, labels]
  )

  const defaultColumns = useMemo<ColumnDef<IDefaultLabelRow, unknown>[]>(
    () => [
      {
        accessorKey: 'key',
        header: 'Label',
        cell: ({ row }) => (
          <LabelBadge
            labelKey={row.original.key}
            labelValue={row.original.value}
            customColor={row.original.color}
            size="sm"
          />
        ),
      },
      {
        id: 'source',
        header: 'Source',
        enableSorting: false,
        cell: () => (
          <Tooltip
            tipContent="Defined in the app config's default_labels"
            position="left"
            tipContentClassName="!whitespace-normal !w-auto max-w-[200px] text-xs"
          >
            <span className="flex items-center gap-2">
              <Icon variant="LockIcon" size="16" />
              <Text variant="subtext" theme="neutral">
                App config
              </Text>
            </span>
          </Tooltip>
        ),
      },
    ],
    []
  )

  const labelColumns = useMemo<ColumnDef<TAppLabelKey, unknown>[]>(
    () => [
      {
        accessorKey: 'key',
        header: 'Label',
        cell: ({ row }) => {
          const values = [...(row.original?.values ?? [])].sort()
          return (
            <LabelBadge
              labelKey={row.original?.key}
              labelValue={values[0] ?? ''}
              customColor={row.original?.color}
              size="sm"
            />
          )
        },
      },
      {
        id: 'values',
        accessorFn: (label) => label?.values?.length ?? 0,
        header: 'Values',
        cell: ({ row }) => {
          const values = [...(row.original?.values ?? [])].sort()
          if (values.length <= 1) {
            return (
              <Text variant="subtext" theme="neutral">
                {values[0] ?? '—'}
              </Text>
            )
          }
          return (
            <span className="flex flex-wrap gap-1">
              {values.map((value) => (
                <Badge key={value} size="sm" theme="default">
                  {value}
                </Badge>
              ))}
            </span>
          )
        },
      },
      {
        id: 'entity-types',
        accessorFn: (label) => label?.entity_types?.join(', ') ?? '',
        header: 'Used on',
        cell: ({ row }) => (
          <span className="flex flex-wrap gap-1">
            {[...(row.original?.entity_types ?? [])].sort().map((entityType) => (
              <Badge key={entityType} size="sm" theme="neutral">
                {entityType}
              </Badge>
            ))}
          </span>
        ),
      },
      {
        accessorKey: 'usage_count',
        header: 'Uses',
        cell: ({ row }) => (
          <Text variant="subtext" family="mono">
            {row.original?.usage_count ?? 0}
          </Text>
        ),
      },
      {
        id: 'color',
        header: 'Color',
        enableSorting: false,
        cell: ({ row }) => (
          <span className="flex items-center justify-end gap-2">
            {row.original?.is_override ? (
              <Badge size="sm" theme="brand">
                Custom
              </Badge>
            ) : null}
            <LabelColorPicker
              id={`label-color-picker-${row.original?.key}`}
              value={row.original?.color}
              defaultColor={row.original?.default_color}
              isOverride={row.original?.is_override}
              disabled={isPending}
              ariaLabel={`Change color for ${row.original?.key}`}
              onSelect={(color) => onOverride(row.original.key, color)}
              onReset={() => onRemoveOverride(row.original.key)}
            />
          </span>
        ),
      },
    ],
    [isPending, onOverride, onRemoveOverride]
  )

  return (
    <>
      <HeadingGroup className="gap-1.5">
        <div className="flex items-center justify-between gap-4">
          <Text variant="h3" weight="stronger" level={1}>
            Labels
          </Text>
          {resetAction}
        </div>
        {defaultRows.length === 0 ? (
          <Text variant="subtext" theme="neutral">
            {OTHER_LABELS_DESCRIPTION}
          </Text>
        ) : null}
      </HeadingGroup>

      {defaultRows.length > 0 ? (
        <>
          <SectionHeader
            title="Default labels"
            description="Applied to every install of this app. Edit them via default_labels in the app config."
          />
          <Table<IDefaultLabelRow>
            columns={defaultColumns}
            data={defaultRows}
            enableSearch={false}
          />
          <SectionHeader
            title="Other labels"
            description={OTHER_LABELS_DESCRIPTION}
          />
        </>
      ) : null}

      <Table<TAppLabelKey>
        columns={labelColumns}
        data={labels}
        enableSearch={false}
        isLoading={isLoading}
        emptyStateProps={{
          variant: 'diagram',
          emptyTitle: 'No labels yet',
          emptyMessage:
            'Add labels to your components, actions, runbooks, or installs to see them here.',
        }}
      />
    </>
  )
}
