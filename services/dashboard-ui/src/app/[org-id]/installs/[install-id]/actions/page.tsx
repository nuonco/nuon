import classNames from 'classnames'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  ActionTriggerButton,
  ActionTriggerType,
  InstallPageSubNav,
  InstallStatuses,
  InstallWorkflowRunHistory,
  DashboardContent,
  Link,
  Section,
  StatusBadge,
  Text,
  Time,
} from '@/components'
import {
  getAppWorkflows,
  getInstall,
  getInstallWorkflowRuns,
  getOrg,
} from '@/lib'

export default withPageAuthRequired(async function InstallWorkflowRuns({
  params,
}) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, install, workflowRuns] = await Promise.all([
    getOrg({ orgId }),
    getInstall({ installId, orgId }),
    getInstallWorkflowRuns({ installId, orgId }),
  ])

  const appWorkflows = await getAppWorkflows({ appId: install.app_id, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        { href: `/${org.id}/installs/${install.id}`, text: install.name },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={<InstallStatuses initInstall={install} shouldPoll />}
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Workflow history">
          <InstallWorkflowRunHistory
            appWorkflows={appWorkflows}
            orgId={orgId}
            installId={installId}
            installWorkflowRuns={workflowRuns}
            shouldPoll
          />
        </Section>

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Action workflows">
            <div className="flex flex-col gap-2 divide-y">
              {appWorkflows.map((aW) => (
                <div
                  key={aW.id}
                  className="flex items-end justify-between py-2 flex-wrap gap-2"
                >
                  <div key={aW.id} className="flex flex-col gap-2">
                    <Link
                      href={`/${orgId}/apps/${install.app_id}/actions/${aW.id}`}
                    >
                      <Text variant="med-12">{aW.name}</Text>
                    </Link>
                    <span className="flex items-center justify-start gap-2 flex-wrap">
                      {aW.configs[0].triggers?.map((t) => (
                        <ActionTriggerType key={t.id} triggerType={t.type} />
                      ))}
                      <Text variant="reg-12">
                        {aW.configs[0].steps.length} Steps
                      </Text>
                    </span>
                  </div>
                  {aW.configs[0].triggers.find((t) => t.type === 'manual') ? (
                    <ActionTriggerButton
                      installId={installId}
                      orgId={orgId}
                      workflowConfigId={aW.configs[0]?.id}
                    />
                  ) : null}
                </div>
              ))}
            </div>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
