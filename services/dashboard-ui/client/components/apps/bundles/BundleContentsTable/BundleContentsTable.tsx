import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import type { TCustomerManagedBundleArtifact } from '@/types'
import { formatBytes } from '@/utils/string-utils'
import { KIND_GROUPS, KindFilter, useSelectedKindGroups } from './KindFilter'

const KIND_LABELS: Record<string, string> = {
  component: 'Component',
  sandbox: 'Sandbox',
  image: 'Image',
  runner_binary: 'Runner binary',
  runner_image: 'Runner image',
  stack_asset: 'Stack asset',
  action_step: 'Action step',
}

const KIND_ORDER = [
  'component',
  'sandbox',
  'image',
  'action_step',
  'runner_binary',
  'runner_image',
  'stack_asset',
]

const shortDigest = (digest?: string) => {
  if (!digest) return '—'
  const value = digest.replace(/^sha256:/, '')
  return `${value.slice(0, 12)}…`
}

const defaultArtifactHref = (
  artifact: TCustomerManagedBundleArtifact,
  orgId?: string,
  appId?: string
) => {
  if (!orgId || !appId) return undefined
  const appPath = `/${orgId}/apps/${appId}`
  switch (artifact.kind) {
    case 'component':
    case 'image':
      return artifact.component_id
        ? `${appPath}/components/${artifact.component_id}`
        : undefined
    case 'sandbox':
      return `${appPath}/sandbox`
    case 'action_step':
      return artifact.action_workflow_id
        ? `${appPath}/actions/${artifact.action_workflow_id}`
        : undefined
    default:
      return undefined
  }
}

export const BundleContentsTable = ({
  artifacts,
  orgId,
  appId,
  getArtifactHref,
}: {
  artifacts: TCustomerManagedBundleArtifact[]
  orgId?: string
  appId?: string
  getArtifactHref?: (
    artifact: TCustomerManagedBundleArtifact
  ) => string | undefined
}) => {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q')?.trim().toLowerCase() ?? ''
  const selectedGroups = useSelectedKindGroups()

  const sorted = useMemo(() => {
    const selectedKinds = new Set(
      KIND_GROUPS.filter((group) => selectedGroups.includes(group.key)).flatMap(
        (group) => group.kinds
      )
    )
    const byKind = artifacts.filter((artifact) =>
      selectedKinds.has(artifact.kind ?? '')
    )
    const filtered = query
      ? byKind.filter((artifact) =>
          [
            artifact.logical_name,
            artifact.repository,
            artifact.kind,
            artifact.digest,
            artifact.media_type,
          ].some((value) => value?.toLowerCase().includes(query))
        )
      : byKind
    return [...filtered].sort((a, b) => {
      const kindDiff =
        KIND_ORDER.indexOf(a.kind ?? '') - KIND_ORDER.indexOf(b.kind ?? '')
      if (kindDiff !== 0) return kindDiff
      return (a.logical_name ?? '').localeCompare(b.logical_name ?? '')
    })
  }, [artifacts, query, selectedGroups])

  const presentGroups = useMemo(() => {
    const kinds = new Set(artifacts.map((artifact) => artifact.kind ?? ''))
    return KIND_GROUPS.filter((group) =>
      group.kinds.some((kind) => kinds.has(kind))
    )
  }, [artifacts])

  const columns: ColumnDef<TCustomerManagedBundleArtifact>[] = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'logical_name',
        cell: (props) => {
          const name = props.getValue<string>() || '—'
          const href =
            getArtifactHref?.(props.row.original) ??
            defaultArtifactHref(props.row.original, orgId, appId)
          return (
            <div className="flex flex-col gap-1">
              {href ? (
                <Link href={href} className="w-fit font-strong">
                  {name}
                </Link>
              ) : (
                <Text variant="body" weight="strong">
                  {name}
                </Text>
              )}
              {props.row.original.repository ? (
                <Text variant="subtext" theme="neutral" className="break-all">
                  {props.row.original.repository}
                </Text>
              ) : null}
            </div>
          )
        },
      },
      {
        header: 'Kind',
        accessorKey: 'kind',
        cell: (props) => {
          const kind = props.getValue<string | undefined>()
          return kind ? (
            <Badge>{KIND_LABELS[kind] ?? kind}</Badge>
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )
        },
      },
      {
        header: 'Digest',
        accessorKey: 'digest',
        cell: (props) => {
          const digest = props.getValue<string | undefined>()
          return digest ? (
            <ID clickToCopyProps={{ copyValue: digest }}>
              {shortDigest(digest)}
            </ID>
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )
        },
      },
      {
        header: 'Config digest',
        accessorKey: 'config_digest',
        cell: (props) => {
          const digest = props.getValue<string | undefined>()
          return digest ? (
            <ID clickToCopyProps={{ copyValue: digest }}>
              {shortDigest(digest)}
            </ID>
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )
        },
      },
      {
        header: 'Media type',
        accessorKey: 'media_type',
        cell: (props) => (
          <Text variant="subtext" theme="neutral" className="break-all">
            {props.getValue<string>() || '—'}
          </Text>
        ),
      },
      {
        header: 'Size',
        accessorKey: 'size',
        cell: (props) => {
          const size = props.getValue<number | undefined>()
          return <Text variant="subtext">{size ? formatBytes(size) : '—'}</Text>
        },
      },
    ],
    [orgId, appId, getArtifactHref]
  )
  const visibleColumns = columns.filter((column) => {
    const accessorKey = (
      column as ColumnDef<TCustomerManagedBundleArtifact> & {
        accessorKey?: string
      }
    ).accessorKey
    if (accessorKey === 'config_digest')
      return artifacts.some(({ config_digest }) => !!config_digest)
    if (accessorKey === 'media_type')
      return artifacts.some(({ media_type }) => !!media_type)
    return true
  })

  return (
    <Table<TCustomerManagedBundleArtifact>
      columns={visibleColumns}
      data={sorted}
      enableSearch
      filterActions={<KindFilter groups={presentGroups} />}
      emptyStateProps={{
        emptyTitle: 'No contents recorded',
        emptyMessage:
          'This bundle has no artifact records. Bundles published before contents tracking will not list artifacts.',
      }}
    />
  )
}
