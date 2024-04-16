import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DateTime } from 'luxon'
import { Code, Heading, Logs, Page, Status, Text } from '@/components'
import { getSandboxRun, getSandboxRunLogs } from '@/lib'
import { sentanceCase } from '@/utils'

export default withPageAuthRequired(
  async function RunDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string
    const runId = params?.['run-id'] as string

    const [run, logs] = await Promise.all([
      getSandboxRun({ installId, orgId, runId }),
      getSandboxRunLogs({ installId, orgId, runId }),
    ])

    return (
      <Page
        heading={
          <div className="flex flex-wrap gap-8 items-end border-b pb-6">
            <div className="flex flex-col flex-auto gap-2">
              <Text variant="overline">{run?.id}</Text>
              <Heading level={1} variant="title">
                {sentanceCase(run?.run_type)}
              </Heading>
              <Text variant="caption">
                Finished {DateTime.fromISO(run?.updated_at).toRelative()}
              </Text>
            </div>

            <div className="flex flex-auto flex-col">
              <Status
                status={run?.status}
                description={run?.status_description}
              />
            </div>
          </div>
        }
        links={[
          { href: orgId },
          { href: installId },
          { href: 'runs/' + runId, text: runId },
        ]}
      >
        <Heading>Sandbox run</Heading>
        <Code variant="preformated">{JSON.stringify(run, null, 2)}</Code>

        <Heading>Sandbox run logs</Heading>
        <Logs logs={logs} />
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
