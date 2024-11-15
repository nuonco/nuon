import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DashboardContent, Section, Text } from '@/components'
import { getApp, getOrg } from '@/lib'

export default withPageAuthRequired(async function AppWorkflow({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const workflowId = params?.['workflow-id'] as string
  const [org, app] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
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
      heading={'Workflow name'}
      headingUnderline={workflowId}
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Jobs">
          <Text>Job list</Text>
        </Section>

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Actions">
            <div className="flex flex-col gap-8">
              <Text>More info</Text>
            </div>
          </Section>

          <Section heading="Config">
            <Text>Even more info</Text>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
