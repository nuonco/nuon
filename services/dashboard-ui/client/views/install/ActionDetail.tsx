import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Duration } from '@/components/common/Duration'
import { ActionStep } from '@/components/actions/ActionStep'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { InstallActionManualRunButton } from '@/components/actions/InstallActionManualRun'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { InstallActionRunTimeline } from '@/components/actions/InstallActionRunTimeline'
import { RemovedFromAppConfigBanner } from '@/components/installs/RemovedFromAppConfig'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import {
  HistoryPanelButton,
  HistoryRail,
} from '@/components/layout/HistoryRail'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getInstallAction, getInstallState } from '@/lib'
import type { TActionConfigTriggerType } from '@/types'
import { sortByIdx } from '@/utils/action-utils'
import { isActionInAppConfig } from '@/utils/app-config-membership'
import { CustomerManagedSnapshotActionDetail } from '@/components/customer-managed-support/SnapshotActions'
import { isCustomerManagedInstall } from '@/utils/install-utils'

export const ActionDetail = () => {
  const { install } = useInstall()

  return isCustomerManagedInstall(install) ? (
    <CustomerManagedSnapshotActionDetail />
  ) : (
    <ConnectedActionDetail />
  )
}

const ConnectedActionDetail = () => {
  const { actionId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: action, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-action', org?.id, install?.id, actionId],
    queryFn: () =>
      getInstallAction({
        orgId: org.id,
        installId: install.id,
        actionId: actionId!,
        limit: 10,
        offset: 0,
      }),
    enabled: !!org?.id && !!install?.id && !!actionId,
    refetchInterval: 20000,
  })

  const { data: installState } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-state', org?.id, install?.id],
    queryFn: () => getInstallState({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const { appConfig } = useInstallAppConfig()
  const removed = !!appConfig && !isActionInAppConfig(appConfig, actionId)

  const installActionBreakGlassRole =
    action?.action_workflow?.configs?.[0]?.break_glass_role_arn
  const breakGlassRoleArns =
    installState?.install_stack?.outputs?.break_glass_role_arns
  const kubeConfigEnabled =
    action?.action_workflow?.configs?.[0]?.enable_kube_config
  const actionImage = action?.action_workflow?.configs?.[0]?.image

  const history = (
    <InstallActionRunTimeline
      actionId={actionId!}
      actionName={action?.action_workflow?.name ?? ''}
      shouldPoll
    />
  )

  const manualTrigger = action?.action_workflow?.configs?.[0]?.triggers?.find(
    (t) => t.type === 'manual'
  )

  return (
    <>
      <PageTitle
        segments={[action?.action_workflow?.name ?? 'Action', install?.name]}
      />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/actions`,
            text: 'Actions',
          },
          {
            path: `/${org?.id}/installs/${install?.id}/actions/${actionId}`,
            text: action?.action_workflow?.name,
          },
        ]}
      />

      <DetailPage
        header={
          <DetailHeader
            title={action?.action_workflow?.name}
            loading={isLoading}
            loadingWidth={20}
            id={action?.action_workflow_id}
            identity={
              action?.id ? (
                <AdminDashboardLink
                  path={`/queues?owner_id=${action.id}`}
                  label="Admin panel"
                />
              ) : null
            }
            actions={
              <>
                <HistoryPanelButton title="Run history" history={history} />
                {manualTrigger ? (
                  removed ? (
                    <Button
                      variant="primary"
                      disabled
                      tooltipProps={{
                        position: 'left',
                        tipContent:
                          "This action is no longer in the install's app config version.",
                      }}
                    >
                      Run action
                      <Icon variant="PlayIcon" />
                    </Button>
                  ) : (
                    <InstallActionManualRunButton
                      action={action!.action_workflow}
                      actionConfigId={action!.action_workflow!.configs![0].id}
                      variant="primary"
                    >
                      Run action
                      <Icon variant="PlayIcon" />
                    </InstallActionManualRunButton>
                  )
                ) : null}
              </>
            }
            metadata={
              isLoading ? (
                <>
                  <LabeledStatus label="Last status" loading />
                  <LabeledValue label="Kube config" loading />
                  <LabeledValue label="Timeout" loading />
                </>
              ) : (
                <>
                  {action?.runs?.[0] ? (
                    <LabeledStatus
                      label="Last status"
                      statusProps={{ status: action.runs[0].status_v2?.status }}
                      tooltipProps={{
                        position: 'top',
                        tipContent:
                          action.runs[0].status_v2?.status_human_description,
                      }}
                    />
                  ) : null}
                  <LabeledValue label="Kube config">
                    <Badge
                      theme={kubeConfigEnabled ? 'info' : 'warn'}
                      variant="code"
                      size="sm"
                    >
                      {kubeConfigEnabled ? 'Enabled' : 'Disabled'}
                    </Badge>
                  </LabeledValue>
                  <LabeledValue label="Timeout">
                    <Duration
                      nanoseconds={
                        action?.action_workflow?.configs?.[0]?.timeout
                      }
                      variant="subtext"
                    />
                  </LabeledValue>
                  {actionImage ? (
                    <LabeledValue label="Container image">
                      <Code variant="inline">{actionImage}</Code>
                    </LabeledValue>
                  ) : null}
                  {action?.runs?.[0] ? (
                    <LabeledValue label="Last trigger">
                      <ActionTriggerType
                        size="sm"
                        triggerType={
                          action.runs[0]
                            .triggered_by_type as TActionConfigTriggerType
                        }
                        componentName={
                          action.runs[0].run_env_vars?.COMPONENT_NAME
                        }
                        componentPath={`/${org?.id}/installs/${install?.id}/components/${action.runs[0].run_env_vars?.COMPONENT_ID}`}
                      />
                    </LabeledValue>
                  ) : null}
                </>
              )
            }
          />
        }
        banners={removed ? <RemovedFromAppConfigBanner kind="action" /> : null}
      >
        <HistoryRail title="Run history" history={history}>
          {installActionBreakGlassRole ? (
            <div className="flex flex-col gap-4">
              <SectionHeader
                title="Break glass role"
                status={
                  <Status
                    status={
                      breakGlassRoleArns?.[installActionBreakGlassRole]
                        ? 'provisioned'
                        : 'not-provisioned'
                    }
                  >
                    {breakGlassRoleArns?.[installActionBreakGlassRole]
                      ? 'Provisioned'
                      : 'Not provisioned'}
                  </Status>
                }
              />
              {breakGlassRoleArns?.[installActionBreakGlassRole] ? (
                <div className="flex flex-col gap-2">
                  <Text variant="body" weight="strong">
                    Role assumed while running this action
                  </Text>
                  <Code variant="default">
                    {breakGlassRoleArns[installActionBreakGlassRole]}
                  </Code>
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  <Text variant="body">
                    Break glass role must be enabled in the install stack before
                    running this action.
                  </Text>
                  <Code variant="default">{installActionBreakGlassRole}</Code>
                </div>
              )}
            </div>
          ) : null}

          {action?.action_workflow?.configs?.[0]?.role ? (
            <div className="flex flex-col gap-2">
              <SectionHeader
                title="Execution role"
                description="IAM role used when executing this action."
              />
              <Code variant="inline">
                {action.action_workflow.configs[0].role}
              </Code>
            </div>
          ) : null}

          <div className="flex flex-col gap-4">
            <SectionHeader title="Steps" />
            {sortByIdx(action?.action_workflow?.configs?.[0]?.steps ?? []).map(
              (step, i) => (
                <ActionStep key={step.id ?? i} index={i} step={step} />
              )
            )}
          </div>
        </HistoryRail>
      </DetailPage>
    </>
  )
}
