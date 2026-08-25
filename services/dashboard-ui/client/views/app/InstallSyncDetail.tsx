import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Navigate, useParams } from 'react-router'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { Badge } from '@/components/common/Badge'
import { BackLink } from '@/components/common/BackLink'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toast } from '@/components/surfaces/Toast'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getAppInstallSync, approveWorkflowStep } from '@/lib'
import type { TAPIError, TWorkflowStep } from '@/types'
import { getStatusTheme } from '@/utils/status-utils'

export const InstallSyncDetail = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const syncId = params.syncId as string
  const hasInstallSyncing = !!org?.features?.['app-install-syncing']

  const { data: sync, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-install-sync', org?.id, app?.id, syncId],
    queryFn: () =>
      getAppInstallSync({ appId: app!.id, syncId, orgId: org!.id }),
    enabled: hasInstallSyncing && !!org?.id && !!app?.id && !!syncId,
    refetchInterval: 5000,
  })

  if (org && !hasInstallSyncing) {
    return <Navigate to={`/${org.id}/apps/${app?.id}`} replace />
  }

  if (isLoading || !sync) {
    return (
      <PageSection>
        <Text variant="body" theme="neutral">
          Loading install sync...
        </Text>
      </PageSection>
    )
  }

  const status = sync.status?.status || 'unknown'
  const statusDescription = sync.status?.status_human_description || ''
  const workflowSteps = sync.workflow?.steps
  const workflowId = sync.workflow?.id

  const approvalStep = workflowSteps?.find(
    (s) =>
      s.execution_type === 'approval' &&
      s.status?.status === 'approval-awaiting' &&
      s.approval?.id &&
      !s.approval?.response
  )

  return (
    <PageSection className="max-w-full">
      <PageTitle segments={['Install sync', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          {
            path: `/${org?.id}/apps/${app?.id}/install-syncs`,
            text: 'Install syncs',
          },
          {
            path: `/${org?.id}/apps/${app?.id}/install-syncs/${syncId}`,
            text: 'Sync',
          },
        ]}
      />

      <BackLink />

      <HeadingGroup className="gap-1.5">
        <div className="flex items-center gap-2.5">
          <Text as="h1" variant="h2" weight="strong" className="leading-tight">
            Install sync
          </Text>
          <Badge size="sm" variant="code">
            {sync.triggered_by}
          </Badge>
        </div>
        <ID className="text-[12px] font-mono text-cool-grey-400 dark:text-cool-grey-500">
          {syncId}
        </ID>
        <div className="flex items-center gap-2 mt-0.5">
          <Status status={status} variant="badge" />
          {statusDescription && (
            <Text variant="subtext" theme="neutral">
              {statusDescription}
            </Text>
          )}
        </div>
      </HeadingGroup>

      {approvalStep && workflowId && (
        <ApprovalBanner
          step={approvalStep}
          workflowId={workflowId}
          orgId={org?.id ?? ''}
          syncQueryKey={['app-install-sync', org?.id, app?.id, syncId]}
        />
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="!p-4 !gap-3">
          <div className="flex items-center gap-2">
            <Icon
              variant="GitBranchIcon"
              size={16}
              className="text-cool-grey-400"
            />
            <Text variant="base" weight="strong">
              Source
            </Text>
          </div>
          {sync.vcs_connection_commit?.sha && (
            <div className="flex items-start gap-3">
              <Icon
                variant="GitCommitIcon"
                size={16}
                className="mt-0.5 shrink-0 text-cool-grey-400"
              />
              <div className="flex flex-col gap-1 min-w-0">
                {sync.vcs_connection_commit.message && (
                  <Text variant="body" weight="strong" className="truncate">
                    {sync.vcs_connection_commit.message
                      .split('\n')[0]
                      ?.trim()}
                  </Text>
                )}
                <div className="flex items-center gap-2 flex-wrap">
                  <Badge size="sm" variant="code">
                    {sync.vcs_connection_commit.sha.slice(0, 8)}
                  </Badge>
                  {sync.vcs_connection_commit.author_name && (
                    <Text variant="subtext" theme="neutral">
                      {sync.vcs_connection_commit.author_name}
                    </Text>
                  )}
                </div>
              </div>
            </div>
          )}
          {!sync.vcs_connection_commit?.sha && (
            <Text variant="subtext" theme="neutral">
              No commit info available
            </Text>
          )}
        </Card>

        <Card className="!p-4 !gap-3">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <Icon
                variant="ClockIcon"
                size={16}
                className="text-cool-grey-400"
              />
              <Text variant="base" weight="strong">
                Details
              </Text>
            </div>
            {sync.queue_id && sync.queue_signal_id && (
              <AdminDashboardLink
                path={`/queues/${sync.queue_id}/signals/${sync.queue_signal_id}`}
                label="View signal"
              />
            )}
          </div>
          <div className="flex flex-col gap-1.5">
            <div className="flex items-center justify-between">
              <Text variant="subtext" theme="neutral">
                Created
              </Text>
              <Time
                variant="subtext"
                time={sync.created_at}
                format="relative"
              />
            </div>
            <div className="flex items-center justify-between">
              <Text variant="subtext" theme="neutral">
                Triggered by
              </Text>
              <Text variant="subtext">{sync.triggered_by}</Text>
            </div>
            {sync.workflow_id && (
              <div className="flex items-center justify-between">
                <Text variant="subtext" theme="neutral">
                  Workflow
                </Text>
                <AdminDashboardLink
                  path={`/workflows/${sync.workflow_id}`}
                  label="View workflow"
                />
              </div>
            )}
          </div>
        </Card>
      </div>

      {workflowSteps && workflowSteps.length > 0 && (
        <div className="flex flex-col gap-3">
          <Text variant="base" weight="strong">
            Steps
          </Text>
          <div className="flex flex-col gap-2">
            {workflowSteps
              .filter(
                (s) =>
                  s.execution_type !== 'hidden' &&
                  s.execution_type !== 'skipped'
              )
              .sort((a, b) => (a.group_idx ?? 0) - (b.group_idx ?? 0))
              .map((step) => {
                const stepStatus = step.status?.status || 'pending'
                const stepTheme = getStatusTheme(stepStatus)
                const meta = step.status?.metadata as
                  | Record<string, unknown>
                  | undefined
                const stepDesc =
                  step.status?.status_human_description ||
                  (meta?.description as string | undefined)
                return (
                  <Card key={step.id} className="!p-3 !gap-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Badge theme={stepTheme} size="sm">
                          {stepStatus.replace(/_/g, ' ').replace(/-/g, ' ')}
                        </Badge>
                        <Text variant="body">{step.name}</Text>
                      </div>
                      {stepDesc && (
                        <Text variant="subtext" theme="neutral">
                          {stepDesc}
                        </Text>
                      )}
                    </div>
                    {step.execution_type === 'approval' &&
                      step.status?.status === 'approval-awaiting' && (
                        <ApprovalPlanView
                          proposedInstalls={
                            (meta?.proposed_installs as Array<{
                              name: string
                              file_path: string
                              config: unknown
                            }>) ?? []
                          }
                        />
                      )}
                    {meta &&
                      Object.keys(meta).length > 0 && (
                        <div className="flex items-center gap-3 flex-wrap">
                          {Object.entries(meta)
                            .filter(
                              ([k, v]) =>
                                k !== 'description' &&
                                v !== null &&
                                v !== undefined &&
                                v !== '' &&
                                typeof v !== 'object'
                            )
                            .map(([k, v]) => (
                              <div key={k} className="flex items-center gap-1">
                                <Text
                                  variant="subtext"
                                  theme="neutral"
                                  className="capitalize"
                                >
                                  {k.replace(/_/g, ' ')}:
                                </Text>
                                <Text variant="subtext">{String(v)}</Text>
                              </div>
                            ))}
                        </div>
                      )}
                  </Card>
                )
              })}
          </div>
        </div>
      )}

      {sync.install_config_syncs && sync.install_config_syncs.length > 0 && (
        <div className="flex flex-col gap-3">
          <Text variant="base" weight="strong">
            Install syncs ({sync.install_config_syncs.length})
          </Text>
          <div className="flex flex-col gap-2">
            {sync.install_config_syncs.map((ics) => (
              <Card key={ics.id} className="!p-3">
                <div className="flex items-center justify-between">
                  <Text variant="body">{ics.install_id}</Text>
                  <div className="flex items-center gap-2">
                    {ics.status?.status_human_description && (
                      <Text variant="subtext" theme="neutral">
                        {ics.status.status_human_description}
                      </Text>
                    )}
                    <Badge
                      theme={getStatusTheme(ics.status?.status || 'unknown')}
                      size="sm"
                    >
                      {ics.status?.status || 'unknown'}
                    </Badge>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}
    </PageSection>
  )
}

const ApprovalPlanView = ({
  proposedInstalls,
}: {
  proposedInstalls: Array<{ name: string; file_path: string; config: unknown }>
}) => {
  if (proposedInstalls.length === 0) return null

  return (
    <div className="flex flex-col gap-2 mt-1">
      {proposedInstalls.map((item) => (
        <div
          key={item.name}
          className="border border-cool-grey-100 dark:border-cool-grey-800 rounded overflow-hidden"
        >
          <div className="flex items-center justify-between px-3 py-2 bg-green-50 dark:bg-green-950 border-b border-cool-grey-100 dark:border-cool-grey-800">
            <div className="flex items-center gap-2">
              <Icon
                variant="PlusCircleIcon"
                size={14}
                className="text-green-600 dark:text-green-400"
              />
              <Text variant="subtext" weight="strong">
                {item.name}
              </Text>
            </div>
            {item.file_path && (
              <Badge size="sm" variant="code">
                {item.file_path}
              </Badge>
            )}
          </div>
          {item.config && (
            <div className="p-3 bg-cool-grey-25 dark:bg-cool-grey-900 font-mono text-xs overflow-x-auto">
              <pre className="whitespace-pre-wrap text-cool-grey-600 dark:text-cool-grey-300">
                {JSON.stringify(item.config, null, 2)}
              </pre>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

const ApprovalBanner = ({
  step,
  workflowId,
  orgId,
  syncQueryKey,
}: {
  step: TWorkflowStep
  workflowId: string
  orgId: string
  syncQueryKey: unknown[]
}) => {
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: respond, isPending } = useMutation({
    mutationFn: (responseType: 'approve' | 'deny-skip-current') =>
      approveWorkflowStep({
        orgId,
        workflowId,
        workflowStepId: step.id ?? '',
        approvalId: step.approval!.id!,
        body: { response_type: responseType, note: '' },
      }),
    onSuccess: (_, responseType) => {
      queryClient.invalidateQueries({ queryKey: syncQueryKey })
      addToast(
        <Toast
          heading={
            responseType === 'approve'
              ? 'Install creation approved'
              : 'Install creation denied'
          }
          theme={responseType === 'approve' ? 'success' : 'info'}
        >
          <Text>
            {responseType === 'approve'
              ? 'Creating installs and continuing sync.'
              : 'Skipped install creation.'}
          </Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Response failed" theme="error">
          <Text>{err?.error || 'Unable to respond to approval.'}</Text>
        </Toast>
      )
    },
  })

  const meta = step.status?.metadata as Record<string, unknown> | undefined
  const missingCount = (meta?.missing_installs as number) ?? 0

  return (
    <Banner theme="warn">
      <div className="flex flex-col gap-3">
        <div className="flex flex-col">
          <Text weight="strong">
            {missingCount} new install{missingCount === 1 ? '' : 's'} need
            {missingCount === 1 ? 's' : ''} approval
          </Text>
          <Text variant="subtext" theme="neutral">
            Review the proposed installs below, then approve to create them or
            deny to skip.
          </Text>
        </div>
        <div className="flex items-center justify-end gap-2">
          <Button
            variant="danger"
            onClick={() => respond('deny-skip-current')}
            disabled={isPending}
          >
            Deny
          </Button>
          <Button
            variant="primary"
            onClick={() => respond('approve')}
            disabled={isPending}
          >
            {isPending ? 'Responding...' : 'Approve'}
          </Button>
        </div>
      </div>
    </Banner>
  )
}
