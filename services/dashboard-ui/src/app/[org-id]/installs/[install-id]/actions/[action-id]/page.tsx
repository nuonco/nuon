import type { Metadata } from 'next'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  ActionTriggerButton,
  ClickToCopy,
  Config,
  ConfigurationVCS,
  ConfigurationVariables,
  Expand,
  InstallWorkflowRunHistory,
  DashboardContent,
  Section,
  StatusBadge,
  Text,
  ToolTip,
  Truncate,
} from '@/components'
import { getInstall, getInstallActionWorkflowRecentRun } from '@/lib'
import { humandReadableTriggeredBy } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const actionWorkflowId = params?.['action-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const [install, action] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallActionWorkflowRecentRun({ actionWorkflowId, installId, orgId }),
  ])

  return {
    title: `${install.name} | ${action.action_workflow?.name}`,
  }
}

export default withPageAuthRequired(async function InstallWorkflowRuns({
  params,
}) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const actionWorkflowId = params?.['action-id'] as string
  const [install, actionWithRecentRuns] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallActionWorkflowRecentRun({ actionWorkflowId, installId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}/actions`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/actions/${actionWorkflowId}`,
          text: actionWithRecentRuns?.action_workflow?.name,
        },
      ]}
      heading={actionWithRecentRuns.action_workflow?.name}
      headingUnderline={actionWithRecentRuns.action_workflow?.id}
      statues={
        <div className="flex gap-6 items-start justify-start">
          {actionWithRecentRuns?.runs?.length ? (
            <>
              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Recent status
                </Text>
                <StatusBadge
                  status={actionWithRecentRuns.runs?.[0]?.status || 'noop'}
                />
              </span>

              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Last trigger
                </Text>
                <Text variant="mono-12">
                  {humandReadableTriggeredBy(
                    actionWithRecentRuns.runs?.[0]?.triggered_by_type
                  )}
                </Text>
              </span>
            </>
          ) : null}
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Install
            </Text>
            <Text variant="med-12">{install.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={install.id}>
                <ClickToCopy>
                  <Truncate variant="small">{install.id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>
          {actionWithRecentRuns?.action_workflow?.configs?.[0]?.triggers?.find(
            (t) => t.type === 'manual'
          ) ? (
            <ActionTriggerButton
              actionWorkflow={actionWithRecentRuns?.action_workflow}
              installId={installId}
              orgId={orgId}
              workflowConfigId={
                actionWithRecentRuns.action_workflow?.configs?.[0]?.id
              }
            />
          ) : null}
        </div>
      }
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Recent executions">
          <InstallWorkflowRunHistory
            orgId={orgId}
            installId={installId}
            actionsWithRecentRuns={actionWithRecentRuns}
            shouldPoll
          />
        </Section>

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Latest configured steps">
            <div className="flex flex-col gap-2">
              {actionWithRecentRuns?.action_workflow?.configs?.[0]?.steps
                ?.sort((a, b) => b?.idx - a?.idx)
                ?.reverse()
                ?.map((s) => {
                  return (
                    <Expand
                      isOpen
                      id={s.id}
                      key={s.id}
                      parentClass="border rounded"
                      headerClass="px-3 py-2"
                      heading={<Text variant="med-12">{s.name}</Text>}
                      expandContent={
                        <div className="flex flex-col gap-4 p-3 border-t">
                          <Config>
                            <ConfigurationVCS vcs={s} />
                          </Config>

                          {s?.env_vars ? (
                            <ConfigurationVariables variables={s.env_vars} />
                          ) : null}
                        </div>
                      }
                    />
                  )
                })}
            </div>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
