import type { Metadata } from 'next'
import { Suspense } from 'react'
import { BackToTop } from '@/components/common/BackToTop'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { WorkflowDetails } from '@/components/workflows/WorkflowDetails'
import { WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'
import { OnboardingCelebrationWrapper } from './OnboardingCelebrationWrapper'
import { getInstallById, getWorkflowById, getOrgById } from '@/lib'
import { snakeToWords, toSentenceCase } from '@/utils/string-utils'
import type { TPageProps } from '@/types'
import { WorkflowSteps, WorkflowStepsError } from './steps'

// NOTE: old layout stuff
import { DashboardContent, Loading, Empty } from '@/components'
import { WorkflowHeader } from '@/components/workflows/WorkflowHeader'
import { WorkflowSteps as OldWorkflowSteps } from './workflow-steps'

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
    getInstallById({ installId, orgId }),
    getWorkflowById({ workflowId: installWorkflowId, orgId }),
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
      getInstallById({ installId, orgId }),
      getWorkflowById({ workflowId: workflowId, orgId }),
      getOrgById({ orgId }),
    ])

  const containerId = 'workflow-page'

  return org?.features?.['stratus-layout'] ? (
    <PageSection id={containerId} isScrollable className="!gap-2 !pb-24">
      <OnboardingCelebrationWrapper>
        {org?.features?.['stratus-workflow'] ? (
          <>
            <WorkflowDetails
              initWorkflow={installWorkflow}
              install={install}
              shouldPoll
            />

            <div className="flex flex-col gap-6 mt-6">
              <Text variant="h3" weight="strong">
                Workflow steps
              </Text>
              <ErrorBoundary fallback={<WorkflowStepsError />}>
                <Suspense fallback={<WorkflowStepsSkeleton />}>
                  <WorkflowSteps
                    orgId={orgId}
                    offset={sp?.['offset'] || '0'}
                    planOnly={installWorkflow?.plan_only}
                    workflowId={workflowId}
                  />
                </Suspense>
              </ErrorBoundary>
            </div>
          </>
        ) : (
          <>
            <WorkflowHeader initWorkflow={installWorkflow} shouldPoll />
            <ErrorBoundary
              fallback={
                <Empty
                  emptyTitle="No workflow steps"
                  emptyMessage="Unable to load workflow steps"
                  variant="404"
                />
              }
            >
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading workflow steps"
                  />
                }
              >
                <OldWorkflowSteps workflowId={workflowId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </>
        )}
      </OnboardingCelebrationWrapper>
      <BackToTop containerId={containerId} />
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/workflows`,
          text: 'Workflows',
        },
        {
          href: `/${orgId}/installs/${install.id}/workflows/${workflowId}`,
          text:
            installWorkflow?.name ||
            snakeToWords(toSentenceCase(installWorkflow?.type)),
        },
      ]}
    >
      <OnboardingCelebrationWrapper>
        <WorkflowHeader initWorkflow={installWorkflow} shouldPoll />
        <ErrorBoundary
          fallback={
            <Empty
              emptyTitle="No workflow steps"
              emptyMessage="Unable to load workflow steps"
              variant="404"
            />
          }
        >
          <Suspense
            fallback={
              <Loading variant="stack" loadingText="Loading workflow steps" />
            }
          >
            <OldWorkflowSteps workflowId={workflowId} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </OnboardingCelebrationWrapper>
    </DashboardContent>
  )
}
