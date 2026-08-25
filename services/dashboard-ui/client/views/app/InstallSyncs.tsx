import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Navigate } from 'react-router'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { ManageInstallsConfigButton } from '@/components/apps/ManageInstallsConfig'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toast } from '@/components/surfaces/Toast'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import {
  getAppInstallSyncs,
  triggerAppInstallSync,
  getAppInstallsConfig,
  approveInstallCreation,
} from '@/lib'
import type { TAPIError, TAppInstallConfigSync } from '@/types'
import { getStatusTheme } from '@/utils/status-utils'

export const InstallSyncs = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const hasInstallSyncing = !!org?.features?.['app-install-syncing']

  const { data: syncs, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-install-syncs', org?.id, app?.id],
    queryFn: () => getAppInstallSyncs({ appId: app!.id, orgId: org!.id }),
    enabled: hasInstallSyncing && !!org?.id && !!app?.id,
    refetchInterval: 10000,
  })

  const { data: installsConfig } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-installs-config', org?.id, app?.id],
    queryFn: () => getAppInstallsConfig({ appId: app!.id, orgId: org!.id }),
    enabled: hasInstallSyncing && !!org?.id && !!app?.id,
  })

  const { mutate: triggerSync, isPending } = useMutation({
    mutationFn: () =>
      triggerAppInstallSync({ appId: app!.id, orgId: org!.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['app-install-syncs', org?.id, app?.id],
      })
      addToast(
        <Toast heading="Sync triggered" theme="info">
          <Text>Syncing install configs for {app?.name}.</Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Sync failed" theme="error">
          <Text>{err?.error || 'Unable to trigger sync.'}</Text>
        </Toast>
      )
    },
  })

  if (org && !hasInstallSyncing) {
    return <Navigate to={`/${org.id}/apps/${app?.id}`} replace />
  }

  return (
    <PageSection>
      <PageTitle title={`Install syncs | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          {
            path: `/${org?.id}/apps/${app?.id}/install-syncs`,
            text: 'Install syncs',
          },
        ]}
      />
      <div className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Install syncs
          </Text>
          <Text variant="subtext" theme="neutral">
            Sync install configurations from git.
          </Text>
        </HeadingGroup>
        <div className="flex items-center gap-3">
          <AdminDashboardLink
            path={`/queues?owner_id=${app?.id}&owner_type=apps&name=app-install-syncs`}
            label="View queue"
          />
          <ManageInstallsConfigButton />
          <Button
            variant="primary"
            onClick={() => triggerSync()}
            disabled={isPending}
          >
            {isPending ? 'Syncing...' : 'Sync now'}
          </Button>
        </div>
      </div>

      {installsConfig && (
        <Card>
          <div className="p-4">
            <Text variant="subtext" weight="strong">
              Config source
            </Text>
            <div className="mt-2 flex items-center gap-3">
              <Badge variant="code" size="md">
                {installsConfig.source === 'config'
                  ? 'installs.toml'
                  : 'dashboard'}
              </Badge>
              <Text variant="subtext">{installsConfig.repo}</Text>
              <Text variant="subtext" theme="neutral">
                {installsConfig.branch}
              </Text>
              {installsConfig.directory !== '.' && (
                <Text variant="subtext" theme="neutral">
                  /{installsConfig.directory}
                </Text>
              )}
            </div>
          </div>
        </Card>
      )}

      {isLoading && (
        <Text variant="subtext" theme="neutral">
          Loading...
        </Text>
      )}

      {!isLoading && (!syncs || syncs.length === 0) && (
        <Card>
          <div className="p-6 text-center">
            <Text variant="subtext" theme="neutral">
              No install syncs yet. Configure an installs config source and
              trigger a sync.
            </Text>
          </div>
        </Card>
      )}

      {syncs?.map((sync) => (
        <SyncCard key={sync.id} sync={sync} />
      ))}
    </PageSection>
  )
}

const SyncCard = ({ sync }: { sync: TAppInstallConfigSync }) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const status = sync.status?.status || 'unknown'
  const theme = getStatusTheme(status)
  const isAwaitingApproval =
    status === 'awaiting_approval' &&
    sync.install_creation_approval?.status === 'pending'
  const approval = sync.install_creation_approval
  const proposedCount = approval?.proposed_installs?.length

  const { mutate: approve, isPending: isApproving } = useMutation({
    mutationFn: (approvalId: string) =>
      approveInstallCreation({
        appId: app!.id,
        syncId: sync.id,
        approvalId,
        orgId: org!.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['app-install-syncs', org?.id, app?.id],
      })
      addToast(
        <Toast heading="Installs approved" theme="success">
          <Text>Creating missing installs and re-running sync.</Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Approval failed" theme="error">
          <Text>{err?.error || 'Unable to approve install creation.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Card>
      <div className="flex items-center justify-between p-4">
        <div className="flex items-center gap-3">
          <Badge theme={theme}>{status.replace(/_/g, ' ')}</Badge>
          <Text variant="subtext">triggered by {sync.triggered_by}</Text>
          {sync.vcs_connection_commit?.sha && (
            <Text variant="subtext" theme="neutral">
              {sync.vcs_connection_commit.sha.slice(0, 7)}
            </Text>
          )}
        </div>
        <div className="flex items-center gap-3">
          {sync.status?.status_human_description && (
            <Text variant="subtext" theme="neutral">
              {sync.status.status_human_description}
            </Text>
          )}
          <Time variant="subtext" time={sync.created_at} format="relative" />
          {sync.queue_id && sync.queue_signal_id && (
            <AdminDashboardLink
              path={`/queues/${sync.queue_id}/signals/${sync.queue_signal_id}`}
              label="View signal"
            />
          )}
          <Link href={`/${org?.id}/apps/${app?.id}/install-syncs/${sync.id}`}>
            View sync
          </Link>
        </div>
      </div>

      {isAwaitingApproval && (
        <div className="border-t px-4 py-3">
          <Banner theme="warn">
            <div className="flex items-center justify-between">
              <Text variant="subtext">
                {typeof proposedCount === 'number'
                  ? `${proposedCount} new install${proposedCount === 1 ? '' : 's'} need${proposedCount === 1 ? 's' : ''} approval`
                  : 'New installs need approval before sync can continue'}
              </Text>
              <Button
                variant="primary"
                size="xs"
                onClick={() => approval?.id && approve(approval.id)}
                disabled={isApproving || !approval?.id}
              >
                {isApproving ? 'Approving...' : 'Approve creation'}
              </Button>
            </div>
          </Banner>
        </div>
      )}

      {sync.install_config_syncs && sync.install_config_syncs.length > 0 && (
        <div className="border-t px-4 py-3">
          <Text variant="subtext" weight="strong">
            {sync.install_config_syncs.length} install
            {sync.install_config_syncs.length === 1 ? '' : 's'}
          </Text>
          <div className="mt-2 flex flex-col gap-1">
            {sync.install_config_syncs.map((ics) => (
              <div key={ics.id} className="flex items-center justify-between">
                <Text variant="subtext">{ics.install_id}</Text>
                <Badge
                  theme={getStatusTheme(ics.status?.status || 'unknown')}
                  size="sm"
                >
                  {ics.status?.status || 'unknown'}
                </Badge>
              </div>
            ))}
          </div>
        </div>
      )}
    </Card>
  )
}
