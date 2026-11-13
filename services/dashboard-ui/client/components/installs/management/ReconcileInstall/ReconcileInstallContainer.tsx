import { useNavigate } from 'react-router'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { extractSections } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import {
  ConfigChangesLoading,
  ConfigChangesViewer,
} from '@/components/branches/ConfigChanges'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useInstallOverview } from '@/hooks/use-install-overview'
import { useInstallLink } from '@/hooks/use-install-path'
import {
  INSTALL_RECONCILIATION_FLAG,
  INSTALL_RECONCILIATION_QUERY_KEY,
  useInstallReconciliation,
} from '@/hooks/use-install-reconciliation'
import { useOrg } from '@/hooks/use-org'
import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { getAppConfigDiff, reconcileInstall } from '@/lib'
import type { TInstallOverviewCommit, TInstallReconciliation } from '@/types'
import { ReconcileInstallModal } from './ReconcileInstall'

export const ReconcileInstallModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const { data: overview } = useInstallOverview()
  const {
    data: reconciliation,
    isLoading,
    refetch,
  } = useInstallReconciliation()
  const target = reconciliation?.target
  const pinned = reconciliation?.pinned

  const {
    data: configDiff,
    isLoading: diffLoading,
    isError: diffError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'app-config-diff',
      org?.id,
      install?.app_id,
      pinned?.app_config_id,
      target?.app_config_id,
    ],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: install!.app_id!,
        configId: target!.app_config_id!,
        oldConfigId: pinned?.app_config_id,
      }),
    enabled:
      !!org?.id &&
      !!install?.app_id &&
      !!pinned?.app_config_id &&
      !!target?.app_config_id,
    retry: 1,
  })
  const diffSections = configDiff?.diff ? extractSections(configDiff.diff) : []
  const versionLabel = target?.version
    ? pinned?.version
      ? `v${pinned.version} → v${target.version}`
      : `v${target.version}`
    : undefined

  const getRunHref = (run?: TInstallOverviewCommit) =>
    org?.id && install?.app_id && run?.branch_id && run?.workflow_id
      ? `/${org.id}/apps/${install.app_id}/branches/${run.branch_id}/runs/${run.workflow_id}`
      : undefined

  const {
    mutate,
    isPending,
    error,
    reset,
  } = useMutation({
    mutationFn: () =>
      reconcileInstall({
        orgId: org!.id,
        installId: install!.id,
        body: { expected_app_config_id: target!.app_config_id! },
      }),
    onSuccess: (result) => {
      const workflowId =
        result?.workflow_id ?? result?.reconciliation?.in_flight_workflow_id
      addToast(
        <Toast heading="Reconciling install" theme="info">
          <Text>
            Reconciling {install?.name} onto{' '}
            {reconciliation?.app_branch_name ?? 'its app branch'}. This may
            take a few minutes.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({
        queryKey: [INSTALL_RECONCILIATION_QUERY_KEY, org?.id, install?.id],
      })
      queryClient.invalidateQueries({ queryKey: ['install', org?.id, install?.id] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      removeModal(props.modalId)
      navigate(
        installLink({
          installId: install!.id,
          appId: install?.app_id,
          suffix: workflowId ? `/workflows/${workflowId}` : '/workflows',
        })
      )
    },
    onError: (err: any) => {
      void refetch()
      addToast(
        <Toast heading="Reconcile failed" theme="error">
          <Text>
            {err?.error || err?.description || 'Unable to reconcile this install.'}
          </Text>
        </Toast>
      )
    },
  })

  return (
    <ReconcileInstallModal
      installName={install?.name ?? ''}
      reconciliation={reconciliation}
      isLoading={isLoading}
      isPending={isPending}
      error={error as any}
      repo={overview?.branch_tracking?.repo}
      getRunHref={getRunHref}
      diff={
        !target?.app_config_id ? undefined : diffLoading && !configDiff ? (
          <ConfigChangesLoading />
        ) : diffError ? (
          <Text variant="subtext" theme="neutral">
            Changes could not be loaded.
          </Text>
        ) : (
          <ConfigChangesViewer
            sections={diffSections}
            versionLabel={versionLabel}
            previousSha={
              pinned?.commit?.sha && pinned.commit.sha !== target.commit?.sha
                ? pinned.commit.sha
                : undefined
            }
            sha={target.commit?.sha}
            message={target.commit?.message}
            author={target.commit?.author}
            createdAt={target.commit?.created_at}
            emptyMessage="This install already runs this config."
          />
        )
      }
      onSubmit={() => {
        reset()
        mutate()
      }}
      {...props}
    />
  )
}

export const reconcileDisabledReason = (
  reconciliation?: TInstallReconciliation
) => {
  switch (reconciliation?.status) {
    case 'behind':
      return undefined
    case 'current':
      return 'Cannot reconcile — install is on the latest config'
    case 'in_progress':
      return 'Cannot reconcile — a reconciliation is already running'
    case 'blocked':
      return `Cannot reconcile — ${
        reconciliation.blockers?.at(0)?.message ?? 'the install is blocked'
      }`
    default:
      return 'Cannot reconcile — reconciliation state is unavailable'
  }
}

export const ReconcileInstallButton = ({
  children,
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const enabled = useOrgFeatureFlag(INSTALL_RECONCILIATION_FLAG)
  const { data: reconciliation } = useInstallReconciliation()

  if (!enabled) return null

  const reason = reconcileDisabledReason(reconciliation)

  return (
    <Button
      disabled={!!reason}
      tooltipProps={reason ? { tipContent: reason } : undefined}
      onClick={() => addModal(<ReconcileInstallModalContainer />)}
      {...props}
    >
      {children ?? (
        <>
          Reconcile install
          <Icon variant="ArrowsClockwiseIcon" />
        </>
      )}
    </Button>
  )
}
