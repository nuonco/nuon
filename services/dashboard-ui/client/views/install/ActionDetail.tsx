import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Code } from '@/components/common/Code'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
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
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Panel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallAction, getInstallState } from '@/lib'
import type { TActionConfigTriggerType } from '@/types'
import { sortByIdx } from '@/utils/action-utils'

export const ActionDetail = () => {
  const { actionId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addPanel } = useSurfaces()

  const { data: action } = useQuery({
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
    queryKey: ['install-state', org?.id, install?.id],
    queryFn: () => getInstallState({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const installActionBreakGlassRole =
    action?.action_workflow?.configs?.[0]?.break_glass_role_arn
  const breakGlassRoleArns =
    installState?.install_stack?.outputs?.break_glass_role_arns
  const kubeConfigEnabled =
    action?.action_workflow?.configs?.[0]?.enable_kube_config

  return (
    <PageSection flush className="flex-1">
      <PageTitle
        title={`${action?.action_workflow?.name ?? 'Action'} | ${install?.name}`}
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

      <div className="@container flex flex-col flex-1">
        <header className="p-6 border-b flex flex-col gap-6">
          <div className="flex flex-wrap items-start gap-4 justify-between w-full">
            <HeadingGroup>
              <BackLink className="mb-4" />
              <Text variant="h3" weight="strong">
                {action?.action_workflow?.name}
              </Text>
              <span className="flex flex-wrap items-center gap-4 mt-1">
                {action?.action_workflow_id ? (
                  <ID>{action.action_workflow_id}</ID>
                ) : null}
                {action?.id ? (
                  <AdminDashboardLink
                    path={`/queues?owner_id=${action.id}`}
                    label="Admin panel"
                  />
                ) : null}
              </span>
            </HeadingGroup>

            <div className="flex items-center gap-4">
              <div className="@5xl:hidden">
                <Button
                  variant="secondary"
                  onClick={() =>
                    addPanel(
                      <Panel heading="Run history">
                        <InstallActionRunTimeline
                          actionId={actionId!}
                          actionName={action?.action_workflow?.name ?? ''}
                          shouldPoll
                        />
                      </Panel>
                    )
                  }
                >
                  <Icon variant="ClockCounterClockwiseIcon" size={16} />
                  Run history
                </Button>
              </div>
              {action?.action_workflow?.configs?.[0]?.triggers?.find(
                (t) => t.type === 'manual'
              ) ? (
                <InstallActionManualRunButton
                  action={action.action_workflow}
                  actionConfigId={action.action_workflow.configs[0].id}
                />
              ) : null}
            </div>
          </div>

          {action?.runs?.[0] ? (
            <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
              <LabeledStatus
                label="Last status"
                statusProps={{ status: action.runs[0].status_v2?.status }}
                tooltipProps={{
                  position: 'top',
                  tipContent:
                    action.runs[0].status_v2?.status_human_description,
                }}
              />
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
                  nanoseconds={action?.action_workflow?.configs?.[0]?.timeout}
                  variant="subtext"
                />
              </LabeledValue>
              <LabeledValue label="Last trigger">
                <ActionTriggerType
                  size="sm"
                  triggerType={
                    action.runs[0].triggered_by_type as TActionConfigTriggerType
                  }
                  componentName={action.runs[0].run_env_vars?.COMPONENT_NAME}
                  componentPath={`/${org?.id}/installs/${install?.id}/components/${action.runs[0].run_env_vars?.COMPONENT_ID}`}
                />
              </LabeledValue>
            </div>
          ) : null}
        </header>

        <div className="grid grid-cols-1 @5xl:grid-cols-12 flex-1">
          <div className="@5xl:col-span-8 flex flex-col gap-6">
            {installActionBreakGlassRole ? (
              <PageSection className="flex flex-col gap-4">
                <div className="flex justify-between items-center gap-4">
                  <Text variant="base" weight="strong">
                    Break glass role
                  </Text>
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
                </div>
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
                      Break Glass Role must be enabled in install stack before
                      running this action.
                    </Text>
                    <Code variant="default">{installActionBreakGlassRole}</Code>
                  </div>
                )}
              </PageSection>
            ) : null}

            {action?.action_workflow?.configs?.[0]?.role ? (
              <PageSection className="flex flex-col gap-2">
                <Text variant="base" weight="strong">
                  Execution role
                </Text>
                <Text variant="subtext">
                  IAM role used when executing this action.
                </Text>
                <Code variant="inline">
                  {action.action_workflow.configs[0].role}
                </Code>
              </PageSection>
            ) : null}

            <PageSection className="flex flex-col gap-4">
              <Text variant="base" weight="strong">
                Steps
              </Text>
              {sortByIdx(action?.action_workflow?.configs?.[0]?.steps ?? []).map((step, i) => (
                <ActionStep key={step.id ?? i} index={i} step={step} />
              )) ?? (
                <Text variant="body" theme="neutral">
                  No steps configured.
                </Text>
              )}
            </PageSection>
          </div>

          <PageSection className="hidden @5xl:flex flex-col @5xl:col-span-4 gap-4">
            <Text variant="base" weight="strong">
              Run history
            </Text>
            <InstallActionRunTimeline
              actionId={actionId!}
              actionName={action?.action_workflow?.name ?? ''}
              shouldPoll
            />
          </PageSection>
        </div>
      </div>

    </PageSection>
  )
}
