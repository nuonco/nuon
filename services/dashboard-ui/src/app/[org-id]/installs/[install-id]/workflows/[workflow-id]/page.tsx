import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DashboardContent, Section, Text } from '@/components'
import { getInstall, getOrg } from '@/lib'

export default withPageAuthRequired(async function InstallWorkflow({ params }) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const workflowId = params?.['workflow-id'] as string
  const [org, install] = await Promise.all([
    getOrg({ orgId }),
    getInstall({ installId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        { href: `/${org.id}/installs/${install.id}/workflows`, text: install.name },
        {
          href: `/${org.id}/installs/${install.id}/workflows/${workflowId}`,
          text: 'workflow name',
        },
      ]}
      heading={'Workflow name'}
      headingUnderline={workflowId}
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Workflow logs">
          <Text>Workflow or job logs here</Text>
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
