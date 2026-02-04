import type { Metadata } from 'next'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Plan } from '@/components/approvals/Plan'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { BackToTop } from '@/components/common/BackToTop'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageSection } from '@/components/layout/PageSection'
import { SandboxHeader } from '@/components/sandbox/SandboxHeader'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { SandboxRunProvider } from '@/providers/sandbox-run-provider'
import { getInstall, getInstallSandboxRun, getWorkflow, getOrg } from '@/lib'
import { toSentenceCase } from '@/utils/string-utils'
import { Logs, LogsError, LogsSkeleton } from './logs'

// NOTE: old layout stuff
import { Section } from '@/components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['run-id']: runId,
  } = await params
  const [{ data: install }, { data: sandboxRun }] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallSandboxRun({ runId, orgId }),
  ])

  return {
    title: `${sandboxRun.run_type} | ${install.name} | Nuon`,
  }
}

export default async function SandboxRuns({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['run-id']: runId,
  } = await params
  const [{ data: install }, { data: sandboxRun }, { data: org }] =
    await Promise.all([
      getInstall({ installId, orgId }),
      getInstallSandboxRun({
        orgId,
        runId,
      }),
      getOrg({ orgId }),
    ])

  const { data: workflow } = await getWorkflow({
    orgId,
    workflowId: sandboxRun?.install_workflow_id,
  })
  const step = workflow
    ? workflow?.steps
        ?.filter((s) => s?.step_target_id === sandboxRun?.id)
        ?.at(-1)
    : null

  const containerId = 'sandbox-run-page'
  return (
    <SandboxRunProvider initSandboxRun={sandboxRun}>
      <SandboxHeader workflow={workflow} stepId={step?.id} />

      <PageSection className="!pb-12" id={containerId} isScrollable>
        <Breadcrumbs
          breadcrumbs={[
            {
              path: `/${orgId}`,
              text: org?.name,
            },
            {
              path: `/${orgId}/installs`,
              text: 'Installs',
            },
            {
              path: `/${orgId}/installs/${installId}`,
              text: install?.name,
            },
            {
              path: `/${orgId}/installs/${installId}/sandbox`,
              text: 'Sandbox',
            },
            {
              path: `/${orgId}/installs/${installId}/sandbox/${runId}`,
              text: toSentenceCase(sandboxRun?.run_type) || 'Run',
            },
          ]}
        />
        {/* old page content */}
        <div>
          {workflow &&
          step &&
          step?.approval &&
          !step?.approval?.response &&
          step?.status?.status !== 'auto-skipped' ? (
            <Section
              className="border-b !px-0 !pt-0"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalBanner step={step} />
              <Plan step={step} />
            </Section>
          ) : null}

          <div>
            <LogStreamProvider
              initLogStream={sandboxRun?.log_stream}
              shouldPoll={sandboxRun?.log_stream?.open}
            >
              <AsyncBoundary
                errorFallback={<LogsError />}
                loadingFallback={<LogsSkeleton />}
              >
                <Logs
                  logStreamId={sandboxRun?.log_stream?.id}
                  logStreamOpen={sandboxRun?.log_stream?.open}
                  orgId={orgId}
                />
              </AsyncBoundary>
            </LogStreamProvider>
          </div>

          {workflow &&
          step &&
          step?.approval &&
          step?.approval?.response &&
          step?.status?.status !== 'auto-skipped' ? (
            <Section
              className="border-t !px-0"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalBanner step={step} />
              <Plan step={step} />
            </Section>
          ) : null}
        </div>
        {/* old page content */}
        <BackToTop containerId={containerId} />
      </PageSection>
    </SandboxRunProvider>
  )
}
