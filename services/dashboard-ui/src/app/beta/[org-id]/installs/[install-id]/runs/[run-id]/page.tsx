import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { FiCloud, FiClock } from 'react-icons/fi'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  Duration,
  Heading,
  Status,
  Text,
  Time,
  ToolTip,
} from '@/components'
import { getInstall, getSandboxRun, getOrg } from '@/lib'
import { sentanceCase } from '@/utils'

export default withPageAuthRequired(
  async function SandboxRuns({ params }) {
    const installId = params?.['install-id'] as string
    const orgId = params?.['org-id'] as string
    const runId = params?.['run-id'] as string
    const [install, org, sandboxRun] = await Promise.all([
      getInstall({ installId, orgId }),
      getOrg({ orgId }),
      getSandboxRun({ installId, orgId, runId }),
    ])

    return (
      <DashboardContent
        breadcrumb={[
          { href: `/beta/${org.id}`, text: org.name },
          { href: `/beta/${org.id}/installs`, text: 'Installs' },
          {
            href: `/beta/${org.id}/installs/${install.id}`,
            text: install.name,
          },
          {
            href: `/beta/${org.id}/installs/${install.id}/runs/${sandboxRun.id}`,
            text: `${install.name} ${sandboxRun.run_type}`,
          },
        ]}
        heading={`${install.name} ${sandboxRun.run_type}`}
        headingUnderline={sandboxRun.id}
        meta={
          <div className="flex gap-8 items-center justify-start pb-6">
            <Text variant="caption">
              <FiCloud />
              <Time time={sandboxRun.created_at} variant="caption" />
            </Text>
            <Text variant="caption">
              <FiClock />
              <Duration
                beginTime={sandboxRun.created_at}
                endTime={sandboxRun.updated_at}
                variant="caption"
              />
            </Text>
          </div>
        }
        statues={
          <div className="flex gap-6 items-start justify-start">
            <span className="flex flex-col gap-2">
              <Text variant="overline">Status:</Text>
              <ToolTip tipContent={sandboxRun.status_description}>
                <Status status={sandboxRun.status} />
              </ToolTip>
            </span>

            <span className="flex flex-col gap-2">
              <Text variant="overline">Type:</Text>
              <Text variant="caption">{sandboxRun.run_type}</Text>
            </span>

            <span className="flex flex-col gap-2">
              <Text variant="overline">Install:</Text>
              <Text variant="label">{install.name}</Text>
              <Text variant="caption">{install.id}</Text>
            </span>
          </div>
        }
      >
        <div className="flex flex-col lg:flex-row flex-auto">
          <section className="flex flex-auto flex-col gap-4 px-6 py-8 border-r overflow-auto">
            <Heading>{sentanceCase(sandboxRun.run_type)} details</Heading>

            <Text>New runner logs here</Text>
          </section>

          <div
            className="divide-y flex flex-col lg:min-w-[550px]
lg:max-w-[550px]"
          >
            <section className="flex flex-col gap-6 px-6 py-8">
              <Heading>Sandbox</Heading>

              <AppSandboxConfig sandboxConfig={sandboxRun.app_sandbox_config} />
              <AppSandboxVariables
                variables={sandboxRun.app_sandbox_config?.variables}
              />
            </section>
          </div>
        </div>
      </DashboardContent>
    )
  },
  { returnTo: '/dashboard' }
)
