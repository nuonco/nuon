import type { Metadata } from 'next'
import { Suspense } from 'react'
import {
  DashboardContent,
  Loading,
  Empty,
} from '@/components'
import { ErrorBoundary } from '@/components/common/ErrorBoundry'
import { WorkflowHeader } from '@/components/workflows/WorkflowHeader'
import { getInstallById, getWorkflowById } from '@/lib'
import { removeSnakeCase, sentanceCase } from '@/utils'
import { WorkflowSteps } from './workflow-steps'

export async function generateMetadata({ params }): Promise<Metadata> {
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
      removeSnakeCase(sentanceCase(installWorkflow?.type))
    }`,
  }
}

export default async function InstallWorkflow({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['workflow-id']: workflowId,
  } = await params
  const [{ data: install }, { data: installWorkflow }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getWorkflowById({ workflowId: workflowId, orgId }),
  ])

  return (
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
            removeSnakeCase(sentanceCase(installWorkflow?.type)),
        },
      ]}
    >
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
              <Loading variant="stack" loadingText="Loading workflow steps" />
            }
          >
            <WorkflowSteps workflowId={workflowId} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </>
    </DashboardContent>
  )
}
