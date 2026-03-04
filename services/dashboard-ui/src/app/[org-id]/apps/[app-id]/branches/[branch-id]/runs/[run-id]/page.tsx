import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Status } from '@/components/common/Status'
import { Time } from '@/components/common/Time'
import { Duration } from '@/components/common/Duration'
import { Card } from '@/components/common/Card'
import { Badge } from '@/components/common/Badge'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { BackLink } from '@/components/common/BackLink'
import { WorkflowDetails } from '@/components/workflows/WorkflowDetails'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { BranchWorkflowCanvas } from '@/app/[org-id]/apps/[app-id]/branches/[branch-id]/canvas/branch-workflow-canvas'
import { getApp, getAppBranch, getOrg, getWorkflow } from '@/lib'
import type { TPageProps } from '@/types'

type TRunPageProps = TPageProps<
  'org-id' | 'app-id' | 'branch-id' | 'run-id'
>

export async function generateMetadata({
  params,
}: TRunPageProps): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['branch-id']: branchId,
    ['run-id']: runId,
  } = await params
  const { data: branch } = await getAppBranch({ appId, branchId, orgId })

  return {
    title: `Run | ${branch?.name || 'Branch'} | Nuon`,
  }
}

export default async function BranchRunDetailPage({
  params,
}: TRunPageProps) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['branch-id']: branchId,
    ['run-id']: runId,
  } = await params

  const [
    { data: workflow, error: workflowError },
    { data: branch },
    { data: app },
    { data: org },
  ] = await Promise.all([
    getWorkflow({ workflowId: runId, orgId }),
    getAppBranch({ appId, branchId, orgId }),
    getApp({ appId, orgId }),
    getOrg({ orgId }),
  ])

  if (workflowError || !workflow) {
    notFound()
  }

  // Get branch run data from workflow
  const branchRun = workflow.app_branch_runs?.[0]
  const status = branchRun?.status || workflow.status || 'unknown'
  const startedAt = branchRun?.started_at || workflow.created_at
  const completedAt = branchRun?.completed_at || workflow.updated_at

  // Calculate duration
  let duration = null
  if (startedAt && completedAt) {
    duration = new Date(completedAt).getTime() - new Date(startedAt).getTime()
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name || '',
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name || '',
          },
          {
            path: `/${orgId}/apps/${appId}/branches`,
            text: 'Branches',
          },
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}`,
            text: branch?.name || '',
          },
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}/runs`,
            text: 'Workflow Runs',
          },
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}/runs/${runId}`,
            text: 'Run Details',
          },
        ]}
      />

      <BackLink className="mb-4">
        Back to workflow runs
      </BackLink>

      {/* Page Header */}
      <div className="flex items-start justify-between mb-6">
        <HeadingGroup>
          <Text variant="h3" weight="stronger">
            Workflow Run
          </Text>
          <ID>{workflow.id}</ID>
          <div className="flex items-center gap-3 mt-2">
            <Status status={status} />
            {branchRun?.app_branch_config?.config_number && (
              <Badge theme="info" size="sm">
                Config v{branchRun.app_branch_config.config_number}
              </Badge>
            )}
          </div>
        </HeadingGroup>
      </div>

      {/* Run Summary Card */}
      <Card className="mb-6">
        <div className="p-6">
          <Text variant="h4" weight="strong" className="mb-4">
            Run Summary
          </Text>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <Text variant="subtext" theme="neutral" className="mb-1">
                Status
              </Text>
              <Status status={status} />
            </div>

            <div>
              <Text variant="subtext" theme="neutral" className="mb-1">
                Started
              </Text>
              {startedAt ? (
                <Time time={startedAt} format="relative" />
              ) : (
                <Text variant="body">Not started</Text>
              )}
            </div>

            <div>
              <Text variant="subtext" theme="neutral" className="mb-1">
                Duration
              </Text>
              {duration ? (
                <Duration nanoseconds={duration * 1e6} />
              ) : (
                <Text variant="body">--</Text>
              )}
            </div>
          </div>

          {branchRun?.error_message && (
            <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
              <Text variant="subtext" theme="neutral" className="mb-2">
                Error Message
              </Text>
              <div className="p-3 bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-800 rounded">
                <Text variant="body" theme="error">
                  {branchRun.error_message}
                </Text>
              </div>
            </div>
          )}

          {branchRun?.app_branch_config && (
            <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
              <Text variant="subtext" theme="neutral" className="mb-2">
                Configuration
              </Text>
              <div className="flex items-center gap-2">
                <Badge theme="info">
                  v{branchRun.app_branch_config.config_number}
                </Badge>
                <Text variant="subtext" theme="neutral">
                  Created{' '}
                  <Time
                    time={branchRun.app_branch_config.created_at}
                    format="relative"
                  />
                </Text>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Workflow Canvas */}
      <div className="mb-6">
        <BranchWorkflowCanvas
          workflow={workflow}
          branchId={branchId}
          appId={appId}
          orgId={orgId}
        />
      </div>

      {/* Workflow Details */}
      {org?.features?.['stratus-workflow'] ? (
        <WorkflowProvider initWorkflow={workflow} shouldPoll>
          <WorkflowDetails />
        </WorkflowProvider>
      ) : (
        <Card>
          <div className="p-6">
            <Text variant="h4" weight="strong" className="mb-4">
              Workflow Details
            </Text>
            <Text variant="body" theme="neutral">
              Detailed workflow information will be displayed here.
            </Text>
          </div>
        </Card>
      )}
    </PageSection>
  )
}