import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Plan } from '@/components/approvals/Plan'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { BackToTop } from '@/components/common/BackToTop'
import { DeployHeader } from '@/components/deploys/DeployHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { DeployProvider } from '@/providers/deploy-provider'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { getComponent, getInstall, getDeploy, getWorkflow, getOrg } from '@/lib'
import { Logs, LogsError, LogsSkeleton } from './logs'

// NOTE: old layout stuff
import { Section } from '@/components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['deploy-id']: deployId,
  } = await params
  const { data: deploy } = await getDeploy({
    deployId,
    installId,
    orgId,
  })

  return {
    title: `${deploy?.install_deploy_type} | ${deploy?.component_name} | Nuon`,
  }
}

export default async function InstallComponentDeploy({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['component-id']: componentId,
    ['deploy-id']: deployId,
  } = await params
  const [
    { data: deploy, error, status },
    { data: install },
    { data: org },
    { data: component },
  ] = await Promise.all([
    getDeploy({
      deployId,
      installId,
      orgId,
    }),
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
    getComponent({ componentId, orgId }),
  ])

  if (error) {
    console.error(
      'Error rendering install deploy page: ',
      `API status: ${status}`,
      error
    )
    if (status === 404) {
      notFound()
    } else {
      // TODO(nnnat): show error message
      notFound()
    }
  }

  const { data: workflow } = await getWorkflow({
    workflowId: deploy?.install_workflow_id,
    orgId,
  })
  const containerId = 'component-deploy-page'
  const step = workflow
    ? workflow?.steps
        ?.filter(
          (s) =>
            s?.step_target_id === deploy?.id && s?.execution_type === 'approval'
        )
        ?.at(-1)
    : null

  return (
    <>
      <DeployProvider initDeploy={deploy}>
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
              path: `/${orgId}/installs/${installId}/components`,
              text: 'Components',
            },
            {
              path: `/${orgId}/installs/${installId}/components/${componentId}`,
              text: deploy?.component_name,
            },
            {
              path: `/${orgId}/installs/${installId}/components/${componentId}/deploys/${deployId}`,
              text: 'Deploy',
            },
          ]}
        />
        <DeployHeader
          component={component}
          workflow={workflow}
          stepId={step?.id}
        />
        <PageSection className="!pb-12" id={containerId} isScrollable>
          <div>
            {workflow &&
            step &&
            step?.approval &&
            !step?.approval?.response &&
            step?.status?.status !== 'auto-skipped' ? (
              <Section
                className="border-b !px-0"
                childrenClassName="flex flex-col gap-6"
                heading="Approve change"
              >
                <ApprovalBanner step={step} />
                <Plan step={step} />
              </Section>
            ) : null}

            <LogStreamProvider
              initLogStream={deploy?.log_stream}
              shouldPoll={deploy?.log_stream?.open}
            >
              <AsyncBoundary
                errorFallback={<LogsError />}
                loadingFallback={<LogsSkeleton />}
              >
                <Logs
                  logStreamId={deploy?.log_stream?.id}
                  orgId={orgId}
                  logStreamOpen={deploy?.log_stream?.open}
                />
              </AsyncBoundary>
            </LogStreamProvider>

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

          {/* old page layout */}
          <BackToTop containerId={containerId} />
        </PageSection>
      </DeployProvider>
    </>
  )
}
