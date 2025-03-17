import cronstrue from 'cronstrue'
import type { Metadata } from 'next'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  ActionTriggerType,
  Config,
  ConfigurationVariables,
  ConfigurationVCS,
  DashboardContent,
  Expand,
  Section,
  Text,
} from '@/components'
import { getApp, getAppActionWorkflow } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const workflowId = params?.['workflow-id'] as string
  const [app, workflow] = await Promise.all([
    getApp({ appId, orgId }),
    getAppActionWorkflow({ orgId, actionWorkflowId: workflowId }),
  ])

  return {
    title: `${app.name} | ${workflow.name}`,
  }
}

export default withPageAuthRequired(async function AppWorkflow({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const workflowId = params?.['workflow-id'] as string
  const [app, workflow] = await Promise.all([
    getApp({ appId, orgId }),
    getAppActionWorkflow({ orgId, actionWorkflowId: workflowId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}/actions`, text: app.name },
        {
          href: `/${orgId}/apps/${app.id}/actions/${workflowId}`,
          text: workflow.name,
        },
      ]}
      heading={workflow.name}
      headingUnderline={workflowId}
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Steps">
          <div className="flex flex-col gap-4">
            {workflow.configs[0].steps
              ?.sort((a, b) => b?.idx - a?.idx)
              ?.reverse()
              ?.map((s, i) => {
                return (
                  <Expand
                    isOpen
                    id={s.id}
                    key={s.id}
                    parentClass="border rounded"
                    headerClass="px-3 py-2"
                    heading={
                      <Text variant="med-12">
                        {i + 1}. {s.name}
                      </Text>
                    }
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

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Triggers">
            <div className="flex flex-col divide-y">
              {workflow.configs[0].triggers.map((t) => (
                <div className="flex gap-2 py-2" key={t.id}>
                  <ActionTriggerType triggerType={t.type} />
                  {t.type === 'cron' ? (
                    <Text variant="reg-12">
                      Will run{' '}
                      {cronstrue
                        .toString(t.cron_schedule, { verbose: true })
                        .toLowerCase()}
                      .
                    </Text>
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
