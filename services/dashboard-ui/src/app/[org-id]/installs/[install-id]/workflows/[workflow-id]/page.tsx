import type { Metadata } from 'next'
import { Suspense } from 'react'
import { BackToTop } from '@/components/common/BackToTop'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { WorkflowDetails } from '@/components/workflows/WorkflowDetails'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'
import { OnboardingCelebrationWrapper } from './OnboardingCelebrationWrapper'
import { getInstall, getWorkflow, getOrg } from '@/lib'
import { snakeToWords, toSentenceCase } from '@/utils/string-utils'
import type { TPageProps } from '@/types'
import { WorkflowSteps, WorkflowStepsError } from './steps'

type TInstallPageProps = TPageProps<'org-id' | 'install-id' | 'workflow-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['workflow-id']: installWorkflowId,
  } = await params
  const [{ data: install }, { data: installWorkflow }] = await Promise.all([
    getInstall({ installId, orgId }),
    getWorkflow({ workflowId: installWorkflowId, orgId }),
  ])

  return {
    title: `${install?.name} | ${
      installWorkflow?.name ||
      snakeToWords(toSentenceCase(installWorkflow?.type))
    }`,
  }
}

export default async function InstallWorkflow({
  params,
  searchParams,
}: TInstallPageProps) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['workflow-id']: workflowId,
  } = await params
  const sp = await searchParams

  const [{ data: install }, { data: installWorkflow }, { data: org }] =
    await Promise.all([
      getInstall({ installId, orgId }),
      getWorkflow({ workflowId: workflowId, orgId }),
      getOrg({ orgId }),
    ])

  const containerId = 'workflow-page'

  return (
    <PageSection id={containerId} isScrollable className="!gap-2 !pb-24">
      <OnboardingCelebrationWrapper>
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
              path: `/${orgId}/installs/${installId}/workflows`,
              text: 'Workflows',
            },
            {
              path: `/${orgId}/installs/${install.id}/workflows/${workflowId}`,
              text:
                installWorkflow?.name ||
                snakeToWords(toSentenceCase(installWorkflow?.type)),
            },
          ]}
        />
        <WorkflowProvider initWorkflow={installWorkflow} shouldPoll>
          <WorkflowDetails />

          <div className="flex flex-col gap-6 mt-6">
            <Text variant="h3" weight="strong">
              Workflow steps
            </Text>
            <ErrorBoundary fallback={<WorkflowStepsError />}>
              <Suspense fallback={<WorkflowStepsSkeleton />}>
                <WorkflowSteps
                  approvalPrompt={installWorkflow?.approval_option === 'prompt'}
                  orgId={orgId}
                  offset={sp?.['offset'] || '0'}
                  planOnly={installWorkflow?.plan_only}
                  workflowId={workflowId}
                />
              </Suspense>
            </ErrorBoundary>
          </div>
        </WorkflowProvider>
      </OnboardingCelebrationWrapper>
      <BackToTop containerId={containerId} />
    </PageSection>
  )
}
