import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { useNavigate, useSearchParams } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallRoleUsages } from '@/lib'
import type { TInstallRoleUsage } from '@/types'
import { IAMRolePoliciesCard, IAMRoleBoundaryExpand } from './IAMRoles'
import type { TInstallRole } from '@/lib/ctl-api/installs/get-latest-install-roles'

const USAGE_LIMIT = 10
const USAGE_OFFSET_PARAM = 'usage_offset'

const principalLabel = (usage: TInstallRoleUsage): string => {
  const job = usage.runner_job
  const metadata = job?.metadata as Record<string, string> | undefined
  switch (job?.owner_type) {
    case 'install_deploys':
      return `Component: ${metadata?.component_name || '—'}`
    case 'install_sandbox_runs':
      return `Sandbox: ${metadata?.sandbox_run_type || 'run'}`
    case 'install_action_workflow_runs':
      return `Action: ${metadata?.action_workflow_name || '—'}`
    default:
      return job?.owner_type || '—'
  }
}

const operationLabel = (usage: TInstallRoleUsage): string => {
  const operation = usage.runner_job?.operation
  if (!operation) return ''
  if (operation === 'apply-plan') return 'apply'
  if (operation.includes('plan')) return 'plan'
  return operation
}

export const InstallRoleDetail = ({
  installRole,
}: {
  installRole: TInstallRole
}) => {
  const { org } = useOrg()
  const { panels, removePanel } = useSurfaces()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  const role = installRole.app_role_config
  const roleName = role?.name ?? ''
  const installId = installRole.install_id ?? ''
  const offset = Number(searchParams.get(USAGE_OFFSET_PARAM) ?? 0)

  const { data: result, isLoading: usagesLoading } = useQuery({
    queryKey: ['install-role-usages', org?.id, installId, roleName, offset],
    queryFn: () =>
      getInstallRoleUsages({
        installId,
        orgId: org.id,
        roleName,
        offset,
        limit: USAGE_LIMIT,
      }),
    enabled: !!org?.id && !!installId && !!roleName,
    placeholderData: keepPreviousData,
  })

  const usages = result?.data

  const columns = useMemo<ColumnDef<TInstallRoleUsage, unknown>[]>(
    () => [
      {
        accessorKey: 'workflow.name',
        header: 'Workflow',
        cell: ({ row }) => (
          <div className="flex flex-col gap-1">
            <Text weight="strong">
              {row.original.workflow?.name ||
                row.original.workflow?.type ||
                'Workflow'}
            </Text>
            <Text variant="subtext" theme="neutral">
              {principalLabel(row.original)}
            </Text>
          </div>
        ),
      },
      {
        id: 'operation',
        header: 'Operation',
        cell: ({ row }) => {
          const op = operationLabel(row.original)
          return op ? (
            <Badge variant="code" size="sm">
              {op}
            </Badge>
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )
        },
      },
      {
        accessorKey: 'runner_job.created_at',
        header: 'When',
        cell: ({ row }) =>
          row.original.runner_job?.created_at ? (
            <Time
              variant="subtext"
              time={row.original.runner_job.created_at}
              format="relative"
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          ),
      },
      {
        id: 'actions',
        header: '',
        enableSorting: false,
        cell: ({ row }) => {
          const workflowId = row.original.workflow?.id
          const stepId = row.original.workflow_step_id
          if (!workflowId) return null
          const href = `/${org?.id}/installs/${installId}/workflows/${workflowId}${
            stepId ? `?panel=${stepId}` : ''
          }`
          return (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                const topPanelId = panels?.at(-1)?.id
                if (topPanelId) removePanel(topPanelId)
                navigate(href)
              }}
            >
              View workflow <Icon variant="CaretRightIcon" size="16" />
            </Button>
          )
        },
      },
    ],
    [org?.id, installId, panels, removePanel, navigate]
  )

  if (!role) return null

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <Text weight="strong">Summary</Text>
        <div className="grid grid-cols-2 gap-6">
          <LabeledValue label="Created at">
            <Time
              variant="subtext"
              time={role.created_at}
              format="long-datetime"
            />
          </LabeledValue>
          <LabeledValue label="Name">{role.name}</LabeledValue>
          <LabeledValue label="Type">
            <Badge variant="code" size="sm">
              {role.type}
            </Badge>
          </LabeledValue>
          <LabeledValue label="Status">
            <Status status={installRole.provisioned ? 'active' : 'inactive'}>
              {installRole.provisioned ? 'Provisioned' : 'Not provisioned'}
            </Status>
          </LabeledValue>
          <LabeledValue label="Last used">
            {installRole.last_used_at ? (
              <Time
                variant="subtext"
                time={installRole.last_used_at}
                format="relative"
              />
            ) : (
              <Text variant="subtext" theme="neutral">
                Never
              </Text>
            )}
          </LabeledValue>
          <LabeledValue label="Cloud ID">
            {installRole.role_id ? (
              <div className="flex items-start gap-1 min-w-0">
                <Text variant="subtext" family="mono" className="break-all">
                  {installRole.role_id}
                </Text>
                <ClickToCopyButton textToCopy={installRole.role_id} />
              </div>
            ) : (
              <Text variant="subtext" theme="neutral">
                —
              </Text>
            )}
          </LabeledValue>
        </div>
      </Card>

      <IAMRolePoliciesCard policies={role.policies} />
      <IAMRoleBoundaryExpand permissionsBoundary={role.permissions_boundary} />

      <Card>
        <div className="flex flex-col gap-2">
          <HeadingGroup>
            <Text weight="strong">Usage</Text>
            <Text variant="subtext" theme="neutral">
              Workflows that used this role.
            </Text>
          </HeadingGroup>
          <Table
            columns={columns}
            data={usages ?? []}
            isLoading={usagesLoading}
            enableSearch={false}
            emptyMessage="This role has not been used yet"
            pagination={{
              hasNext: result?.pagination?.hasNext ?? false,
              offset,
              limit: USAGE_LIMIT,
              param: USAGE_OFFSET_PARAM,
            }}
          />
        </div>
      </Card>
    </div>
  )
}
