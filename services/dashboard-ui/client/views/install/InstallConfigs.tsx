import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getInstallConfigVersions, triggerInstallConfigSync } from '@/lib'
import type { TInstallConfigVersion } from '@/types'

export const InstallConfigs = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { data: versions, isLoading } = useQuery({
    queryKey: ['install-config-versions', org?.id, install?.id],
    queryFn: () =>
      getInstallConfigVersions({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const { mutate: syncNow, isPending: isSyncing } = useMutation({
    mutationFn: () =>
      triggerInstallConfigSync({
        appId: install!.app_id!,
        branchId: install!.app_branch_id!,
        orgId: org!.id,
        installName: install?.name,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['install-config-versions', org?.id, install?.id],
      })
      addToast(
        <Toast heading="Config sync triggered" theme="info">
          <Text>Syncing configs for {install?.name}.</Text>
        </Toast>
      )
    },
    onError: (err) => {
      addToast(
        <Toast heading="Config sync failed" theme="error">
          <Text>{err?.error || 'Unable to trigger config sync.'}</Text>
        </Toast>
      )
    },
  })

  const canSync = !!install?.app_branch_id

  return (
    <PageSection>
      <PageTitle title={`Configs | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/configs`,
            text: 'Configs',
          },
        ]}
      />

      <div className="flex items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Config history
          </Text>
          <Text variant="subtext" theme="neutral">
            Install config versions synced from git.
          </Text>
        </HeadingGroup>
        {canSync && (
          <div className="shrink-0">
            <Button
              variant="secondary"
              size="sm"
              disabled={isSyncing}
              onClick={() => syncNow()}
            >
              {isSyncing ? 'Syncing...' : 'Sync now'}
            </Button>
          </div>
        )}
      </div>

      {isLoading && (
        <Card>
          <Text variant="subtext" theme="neutral">
            Loading...
          </Text>
        </Card>
      )}

      {!isLoading && (!versions || versions.length === 0) && (
        <Card>
          <EmptyState
            emptyTitle="No configs yet"
            emptyMessage="Config versions will appear here after the first sync from git."
            variant="table"
          />
        </Card>
      )}

      {versions && versions.length > 0 && (
        <div className="flex flex-col gap-3">
          {versions.map((version) => (
            <ConfigVersionRow key={version.id} version={version} />
          ))}
        </div>
      )}
    </PageSection>
  )
}

const ConfigVersionRow = ({ version }: { version: TInstallConfigVersion }) => {
  const triggeredBy = version.metadata?.triggered_by ?? 'unknown'

  return (
    <Card>
      <div className="flex items-center gap-3 flex-wrap">
        <Status
          status={version.status?.status ?? 'unknown'}
          variant="badge"
        />
        <Badge size="sm" theme="neutral">
          {triggeredBy}
        </Badge>
        <Badge size="sm" theme={version.created ? 'success' : 'info'}>
          {version.created ? 'Created' : 'Updated'}
        </Badge>
        {version.file_path && (
          <Text variant="subtext" theme="neutral" className="font-mono text-xs">
            {version.file_path}
          </Text>
        )}
        <div className="ml-auto flex items-center gap-3">
          {version.created_at && (
            <Time
              variant="subtext"
              theme="neutral"
              time={version.created_at}
              format="relative"
            />
          )}
        </div>
      </div>
    </Card>
  )
}
