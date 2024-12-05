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
import { getApp, getOrg, getWorkflow } from '@/lib'

export default withPageAuthRequired(async function AppWorkflow({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const workflowId = params?.['workflow-id'] as string
  const [org, app, workflow] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
    getWorkflow({ orgId, workflowId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/apps`, text: 'Apps' },
        { href: `/${org.id}/apps/${app.id}/workflows`, text: app.name },
        {
          href: `/${org.id}/apps/${app.id}/workflows/${workflowId}`,
          text: 'workflow name',
        },
      ]}
      heading={workflow.name}
      headingUnderline={workflowId}
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Steps">
          <div className="flex flex-col gap-4">
            {workflow.configs[0].steps.map((s, i) => {
              return (
                <Expand
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

                      <ConfigurationVariables variables={s.env_vars} />
                    </div>
                  }
                />
              )
            })}
          </div>
        </Section>

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Triggers">
            <div className="flex gap-2">
              {workflow.configs[0].triggers.map((t) => (
                <ActionTriggerType key={t.id} triggerType={t.type} />
              ))}
            </div>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
