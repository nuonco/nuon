import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { Badge } from '@/components/common/Badge'
import { Code } from '@/components/common/Code'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Duration } from '@/components/common/Duration'
import { StatusWithDescription } from '@/components/common/StatusWithDescription'
import { ActionStep } from '@/components/actions/ActionStep'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { InstallActionManualRunButton } from '@/components/actions/InstallActionManualRun'
import { InstallActionRunTimeline } from '@/components/actions/InstallActionRunTimeline'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallAction, getInstallState } from '@/lib'
import type { TActionConfigTriggerType } from '@/types'

const CONTAINER_ID = 'install-action-detail-page'

export const ActionDetail = () => {
  const { actionId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

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
    <PageSection id={CONTAINER_ID} isScrollable className="!p-0 !gap-0">
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

      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {action?.action_workflow?.name}
          </Text>
          {action?.action_workflow_id ? (
            <ID>{action.action_workflow_id}</ID>
          ) : null}
        </HeadingGroup>
        <div className="flex items-center gap-6">
          {action?.runs?.[0] ? (
            <>
              <LabeledValue label="Last status">
                <StatusWithDescription
                  statusProps={{ status: action.runs[0].status_v2?.status }}
                  tooltipProps={{
                    position: 'top',
                    tipContent:
                      action.runs[0].status_v2?.status_human_description,
                  }}
                />
              </LabeledValue>
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
            </>
          ) : null}
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

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <PageSection className="md:col-span-8">
          <Text variant="base" weight="strong">
            Run history
          </Text>
          <InstallActionRunTimeline
            actionId={actionId!}
            actionName={action?.action_workflow?.name ?? ''}
            shouldPoll
          />
        </PageSection>

        <div className="md:col-span-4 divide-y flex flex-col">
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
            {action?.action_workflow?.configs?.[0]?.steps?.map((step, i) => (
              <ActionStep key={step.id ?? i} index={i} step={step} />
            )) ?? (
              <Text variant="body" theme="neutral">
                No steps configured.
              </Text>
            )}
          </PageSection>
        </div>
      </div>

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
