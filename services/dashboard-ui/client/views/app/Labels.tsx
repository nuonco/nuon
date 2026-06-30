import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { EditLabelColors } from '@/components/apps/EditLabelColors'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Skeleton } from '@/components/common/Skeleton'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/common/Toast'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { getAppLabels, updateApp } from '@/lib'
import type { TAppLabelKey } from '@/lib/ctl-api/apps/get-app-labels'
import type { TAPIError } from '@/types/dashboard.types'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'

type LabelRow = {
  key: string
  preview: ReactNode
  color: string
  isOverride: boolean
  colorCell: ReactNode
  values: ReactNode
  entityTypes: ReactNode
  usageCount: number
}

function parseLabelRows(labels: TAppLabelKey[]): LabelRow[] {
  return labels.map((lk) => ({
    key: lk.key,
    preview: (
      <LabelBadge
        labelKey={lk.key}
        labelValue={lk.values?.[0] ?? ''}
        customColor={lk.color}
        size="sm"
        variant="code"
      />
    ),
    color: lk.color,
    isOverride: lk.is_override,
    colorCell: (
      <span className="flex items-center gap-2">
        <span
          className="inline-block w-4 h-4 rounded border border-cool-grey-300 dark:border-dark-grey-500"
          style={{ backgroundColor: lk.color }}
        />
        <Text variant="subtext" className="font-mono">{lk.color}</Text>
        {lk.is_override && <Badge size="sm" theme="brand">override</Badge>}
      </span>
    ),
    values: (
      <span className="flex flex-wrap gap-1">
        {lk.values?.sort().map((v) => (
          <Badge key={v} size="sm" theme="default">{v}</Badge>
        ))}
      </span>
    ),
    entityTypes: (
      <span className="flex flex-wrap gap-1">
        {lk.entity_types?.sort().map((et) => (
          <Badge key={et} size="sm" theme="info">{et}</Badge>
        ))}
      </span>
    ),
    usageCount: lk.usage_count,
  }))
}

const columns: ColumnDef<LabelRow>[] = [
  {
    accessorKey: 'key',
    header: 'Label key',
    cell: (info) => <Text variant="body" weight="strong">{info.getValue() as string}</Text>,
    enableSorting: true,
  },
  {
    accessorKey: 'preview',
    header: 'Preview',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'colorCell',
    header: 'Color',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'values',
    header: 'Values',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'entityTypes',
    header: 'Used in',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'usageCount',
    header: 'Usage',
    cell: (info) => <Text variant="subtext">{info.getValue() as number}</Text>,
    enableSorting: true,
  },
]

export const Labels = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const { addModal } = useSurfaces()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['app-labels', org?.id, app?.id],
    queryFn: () => getAppLabels({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const { mutate: resetColors, isPending } = useMutation({
    mutationFn: () =>
      updateApp({
        appId: app.id,
        orgId: org.id,
        body: { label_colors: {} },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-labels', org?.id, app?.id] })
      queryClient.invalidateQueries({ queryKey: ['app', org?.id, app?.id] })
      addToast(
        <Toast heading="Label colors reset" theme="success">
          <Text>All label colors reset to defaults.</Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Reset failed" theme="error">
          <Text>{err?.error || 'Unable to reset label colors.'}</Text>
        </Toast>
      )
    },
  })

  const hasOverrides = data?.labels?.some((l) => l.is_override) ?? false
  const rows = parseLabelRows(data?.labels ?? [])

  return (
    <PageSection>
      <PageTitle title={`Labels | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/labels`, text: 'Labels' },
        ]}
      />

      <HeadingGroup>
        <Text variant="base" weight="strong">
          Labels
        </Text>
        <Text variant="subtext" theme="neutral">
          All label keys used across components, actions, runbooks, and installs. Colors are assigned automatically and can be overridden.
        </Text>
      </HeadingGroup>

      {isLoading ? (
        <Card className="flex flex-col gap-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} height="32px" width="100%" />
          ))}
        </Card>
      ) : rows.length === 0 ? (
        <EmptyState
          variant="diagram"
          emptyTitle="No labels yet"
          emptyMessage="Add labels to your components, actions, runbooks, or installs to see them here."
        />
      ) : (
        <Table<LabelRow>
          columns={columns}
          data={rows}
          isLoading={false}
          filterActions={
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => addModal(<EditLabelColors />)}
              >
                <Icon variant="PencilSimpleIcon" size="16" />
                Edit colors
              </Button>
              {hasOverrides && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => resetColors()}
                  disabled={isPending}
                >
                  <Icon variant="ArrowCounterClockwiseIcon" size="16" />
                  Reset to defaults
                </Button>
              )}
            </div>
          }
          pagination={{ limit: 100, offset: 0 }}
        />
      )}
    </PageSection>
  )
}
