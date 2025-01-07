import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  ActionTriggerButton,
  InstallWorkflowRunHistory,
  DashboardContent,
  Section,
  StatusBadge,
  Text,
} from '@/components'
import { getInstall, getInstallActionWorkflowRecentRun, getOrg } from '@/lib'

export default withPageAuthRequired(async function InstallWorkflowRuns({
  params,
}) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const actionWorkflowId = params?.['action-id'] as string
  const [org, install, actionWithRecentRuns] = await Promise.all([
    getOrg({ orgId }),
    getInstall({ installId, orgId }),
    getInstallActionWorkflowRecentRun({ actionWorkflowId, installId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}/actions`,
          text: install.name,
        },
        {
          href: `/${org.id}/installs/${install.id}/actions/${actionWorkflowId}`,
          text: actionWithRecentRuns?.action_workflow?.name,
        },
      ]}
      heading={actionWithRecentRuns.action_workflow?.name}
      headingUnderline={actionWithRecentRuns.action_workflow?.id}
      statues={
        <div className="flex gap-6 items-start justify-start">
          {actionWithRecentRuns?.recent_runs ? (
            <>
              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Recent status
                </Text>
                <StatusBadge
                  status={
                    actionWithRecentRuns.recent_runs?.[0]?.status || 'noop'
                  }
                />
              </span>

              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Last trigger
                </Text>
                <Text variant="mono-12">
                  {actionWithRecentRuns.recent_runs?.[0]?.trigger_type}
                </Text>
              </span>
            </>
          ) : null}
          {actionWithRecentRuns?.action_workflow?.configs?.[0]?.triggers?.find(
            (t) => t.type === 'manual'
          ) ? (
            <ActionTriggerButton
              variant="primary"
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
          <Section className="flex-initial" heading="Workflow steps">
            <div className="flex flex-col gap-2 divide-y">Steps info here</div>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
