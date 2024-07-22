import { GoClock, GoCloud } from 'react-icons/go'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Duration,
  Grid,
  Heading,
  InstallCloudPlatformDetailsCard,
  InstallSandboxDetailsCard,
  InstallSandboxRunLogsCard,
  InstallSandboxRunStatus,
  Page,
  PageHeader,
  PageTitle,
  PageSummary,
  Text,
  Time,
} from '@/components'
import { InstallProvider, SandboxRunProvider } from '@/context'
import { getSandboxRun, getSandboxRunLogs, getInstall, getOrg } from '@/lib'
import type { TSandboxRunLogs } from '@/types'

export default withPageAuthRequired(
  async function RunDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string
    const runId = params?.['run-id'] as string

    const [run, logs, install, org] = await Promise.all([
      getSandboxRun({ installId, orgId, runId }),
      getSandboxRunLogs({ installId, orgId, runId }).catch(console.error),
      getInstall({ installId, orgId }),
      getOrg({ orgId }),
    ])

    return (
      <InstallProvider initInstall={install}>
        <SandboxRunProvider initRun={run} shouldPoll>
          <Page
            header={
              <PageHeader
                info={
                  <>
                    <InstallSandboxRunStatus />
                    <div className="flex flex-col flex-auto gap-1">
                      <Text variant="caption">
                        <b>Run ID:</b> {run.id}
                      </Text>
                      <Text variant="caption">
                        <b>Install ID:</b> {run.install_id}
                      </Text>
                    </div>
                  </>
                }
                title={
                  <PageTitle
                    overline={run.id}
                    title={`${install.name} ${run.run_type}`}
                  />
                }
                summary={
                  <PageSummary>
                    <Text variant="caption">
                      <GoCloud />
                      <Time time={run.updated_at} />
                    </Text>
                    <Text variant="caption">
                      <GoClock />
                      <Duration
                        unitDisplay="short"
                        listStyle="long"
                        variant="caption"
                        beginTime={run.created_at}
                        endTime={run.updated_at}
                      />
                    </Text>
                  </PageSummary>
                }
              />
            }
            links={[
              { href: orgId, text: org?.name },
              { href: installId, text: install?.name },
              { href: 'runs/' + runId, text: runId },
            ]}
          >
            <Grid variant="3-cols">
              <div className="flex flex-col gap-6">
                <Heading variant="subtitle">Install details</Heading>
                <InstallSandboxDetailsCard />
                <InstallCloudPlatformDetailsCard />
              </div>

              <div className="flex flex-col gap-6 lg:col-span-2">
                <Heading variant="subtitle">Run details</Heading>
                <InstallSandboxRunLogsCard
                  initLogs={logs as TSandboxRunLogs}
                  shouldPoll
                />
              </div>
            </Grid>
          </Page>
        </SandboxRunProvider>
      </InstallProvider>
    )
  },
  { returnTo: '/dashboard' }
)
