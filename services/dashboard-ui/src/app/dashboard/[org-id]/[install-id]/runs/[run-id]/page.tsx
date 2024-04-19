import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DateTime } from 'luxon'
import {
  Card,
  CloudDetails,
  Code,
  Grid,
  Heading,
  Logs,
  Page,
  PageHeader,
  SandboxDetails,
  Status,
  Text,
} from '@/components'
import { getSandboxRun, getSandboxRunLogs, getInstall } from '@/lib'
import { sentanceCase } from '@/utils'

export default withPageAuthRequired(
  async function RunDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string
    const runId = params?.['run-id'] as string

    const [run, logs, install] = await Promise.all([
      getSandboxRun({ installId, orgId, runId }),
      getSandboxRunLogs({ installId, orgId, runId }),
      getInstall({ installId, orgId }),
    ])

    return (
      <Page
        header={
          <PageHeader
            info={
              <Status
                status={run?.status}
                description={run?.status_description}
              />
            }
            title={
              <span className="flex flex-col flex-auto gap-2">
                <Text variant="overline">{run?.id}</Text>
                <Heading level={1} variant="title">
                  {sentanceCase(run?.run_type)}
                </Heading>
              </span>
            }
            summary={
              <Text variant="caption">
                Finished {DateTime.fromISO(run?.updated_at).toRelative()}
              </Text>
            }
          />
        }
        links={[
          { href: orgId },
          { href: installId },
          { href: 'runs/' + runId, text: runId },
        ]}
      >
        <Grid variant="3-cols">
          <div className="flex flex-col gap-6">
            <Heading variant="subtitle">Install details</Heading>
            <Card>
              <SandboxDetails {...run?.app_sandbox_config} />
            </Card>

            <Card>
              <CloudDetails {...install} />
            </Card>
          </div>

          <div className="flex flex-col gap-6 lg:col-span-2">
            <Heading variant="subtitle">Run details</Heading>
            <Card>
              <Heading>Run logs</Heading>
              <Logs logs={logs} />
            </Card>
          </div>
        </Grid>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
