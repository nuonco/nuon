import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Badge } from '@/components/common/Badge'
import { Time } from '@/components/common/Time'
import { Card } from '@/components/common/Card'
import { Loading } from '@/components/common/Loading'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { getBranchWorkflowRuns } from '@/lib'
import { useEffect, useState } from 'react'
import type { TInstallWorkflow, TInstallWorkflowStep } from '@/types'

export const BranchRunDetail = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string
  const runId = params.runId as string
  const [selectedStep, setSelectedStep] = useState<TInstallWorkflowStep | null>(null)

  const { data: runs = [], isLoading: isLoadingRuns } = useQuery({
    queryKey: ['branch-runs', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId,
        branchId,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
  })

  const run = runs.find((r) => r.id === runId)
  const steps = run?.steps || []

  useEffect(() => {
    if (steps.length > 0 && !selectedStep) {
      const inProgressStep = steps.find(
        (step) => step.status?.status === 'in-progress'
      )
      setSelectedStep(inProgressStep || steps[0])
    }
  }, [steps, selectedStep])

  if (isLoadingRuns || !run) {
    return (
      <PageSection>
        <Text variant="body" theme="neutral">
          Loading workflow run...
        </Text>
      </PageSection>
    )
  }

  const status = run.status?.status || 'unknown'
  const statusDescription = run.status?.status_human_description || ''

  return (
    <PageSection className="max-w-full">
      <PageTitle title={`Run | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}`, text: branchId },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}/runs`, text: 'Runs' },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}/runs/${runId}`, text: runId },
        ]}
      />
      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="strong">
            Workflow Run
          </Text>
          <ID>{runId}</ID>
          <div className="flex items-center gap-3 mt-2">
            <Badge
              theme={
                status === 'success'
                  ? 'success'
                  : status === 'error'
                  ? 'error'
                  : status === 'in-progress'
                  ? 'info'
                  : 'neutral'
              }
              size="sm"
            >
              {status}
            </Badge>
            {statusDescription && (
              <Text variant="subtext" theme="neutral">
                {statusDescription}
              </Text>
            )}
          </div>
        </HeadingGroup>
        <div className="flex flex-col items-end gap-1">
          <Text variant="subtext" theme="neutral">
            Created <Time time={run.created_at} format="relative" />
          </Text>
          {run.started_at && (
            <Text variant="subtext" theme="neutral">
              Started <Time time={run.started_at} format="relative" />
            </Text>
          )}
          {run.finished_at && (
            <Text variant="subtext" theme="neutral">
              Finished <Time time={run.finished_at} format="relative" />
            </Text>
          )}
        </div>
      </div>

      <Card>
        <div className="p-6 min-w-0">
          <div className="flex items-center justify-between mb-4">
            <Text variant="h3" weight="strong">
              Workflow Progress
            </Text>
            <Text variant="subtext" theme="neutral">
              Scroll horizontally or use trackpad to navigate
            </Text>
          </div>

          {steps.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 gap-4">
              <Loading variant="large" />
              <Text variant="body" theme="neutral">
                Generating workflow steps...
              </Text>
            </div>
          ) : (
            <div
              className="relative overflow-x-auto overflow-y-hidden"
              style={{
                scrollbarWidth: 'thin',
                scrollBehavior: 'smooth',
              }}
            >
              <div className="flex items-center gap-6 py-6 px-4 min-w-max">
                {steps.map((step, idx) => {
                  const stepStatus = step.status?.status || 'pending'
                  const isInProgress = stepStatus === 'in-progress'
                  const isSuccess = stepStatus === 'success'
                  const isError = stepStatus === 'error'
                  const isPending = !isInProgress && !isSuccess && !isError

                  return (
                    <div key={step.id || idx} className="flex items-center gap-4">
                      <div
                        className={`flex flex-col items-center min-w-[240px] p-8 rounded-lg transition-all cursor-pointer border-2 ${
                          selectedStep?.id === step.id
                            ? 'ring-2 ring-primary-300 dark:ring-primary-700 shadow-2xl scale-105 bg-primary-50 dark:bg-dark-grey-900 border-primary-200 dark:border-primary-400/50'
                            : isInProgress
                            ? 'ring-2 ring-blue-200 dark:ring-blue-800 shadow-xl hover:shadow-2xl bg-blue-50 dark:bg-dark-grey-900 border-blue-400 dark:border-blue-500/40'
                            : isSuccess
                            ? 'shadow-lg hover:shadow-xl bg-green-50 dark:bg-dark-grey-900 border-green-400 dark:border-green-500/40'
                            : isError
                            ? 'shadow-lg hover:shadow-xl bg-red-50 dark:bg-dark-grey-900 border-red-300 dark:border-red-500/40'
                            : 'border-dashed border-cool-grey-300 dark:border-dark-grey-600 hover:border-solid hover:shadow-md bg-cool-grey-50 dark:bg-dark-grey-900'
                        }`}
                        onClick={() => setSelectedStep(step)}
                      >
                        <div
                          className={`w-16 h-16 rounded-full flex items-center justify-center mb-4 transition-all ${
                            isInProgress
                              ? 'bg-blue-500 dark:bg-blue-600 text-white shadow-lg'
                              : isSuccess
                              ? 'bg-green-500 dark:bg-green-600 text-white shadow-md'
                              : isError
                              ? 'bg-red-500 dark:bg-red-600 text-white shadow-md'
                              : 'bg-cool-grey-300 dark:bg-dark-grey-400 text-cool-grey-600 dark:text-dark-grey-200'
                          }`}
                        >
                          {isInProgress ? (
                            <Icon variant="Play" size={32} />
                          ) : isSuccess ? (
                            <Icon variant="Check" size={32} />
                          ) : isError ? (
                            <Icon variant="X" size={32} />
                          ) : (
                            <Icon variant="Clock" size={28} />
                          )}
                        </div>

                        <Text variant="base" weight="stronger" className="text-center mb-2">
                          Step {idx + 1}
                        </Text>
                        <Text variant="base" theme="neutral" className="text-center mb-3 max-w-[200px]">
                          {step.name || 'Unknown'}
                        </Text>

                        <div className="flex flex-col gap-2 items-center w-full">
                          {step.group_idx !== undefined && (
                            <Badge
                              theme={
                                isInProgress ? 'info'
                                : isSuccess ? 'success'
                                : isError ? 'error'
                                : 'neutral'
                              }
                              size="md"
                            >
                              Group {step.group_idx}
                            </Badge>
                          )}
                          {step.execution_time && (
                            <Text variant="base" theme="neutral" family="mono" weight="strong">
                              {(step.execution_time / 1000000000).toFixed(1)}s
                            </Text>
                          )}
                        </div>
                      </div>

                      {idx < steps.length - 1 && (
                        <div className="flex items-center">
                          <Icon
                            variant="ArrowRight"
                            size={36}
                            className={`transition-colors ${
                              isSuccess
                                ? 'text-green-500 dark:text-green-400'
                                : 'text-cool-grey-400 dark:text-dark-grey-500'
                            }`}
                          />
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
      </Card>

      {selectedStep && (
        <Card>
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <Text variant="h3" weight="strong">
                Step Details
              </Text>
              <Button variant="ghost" size="sm" onClick={() => setSelectedStep(null)}>
                <Icon variant="X" size={20} />
              </Button>
            </div>

            <div className="space-y-4">
              <div className="flex items-start justify-between">
                <div>
                  <Text variant="base" weight="strong" className="mb-2">
                    {selectedStep.name || 'Unknown step'}
                  </Text>
                  <div className="flex items-center gap-3">
                    <Badge
                      theme={
                        selectedStep.status?.status === 'success'
                          ? 'success'
                          : selectedStep.status?.status === 'error'
                          ? 'error'
                          : selectedStep.status?.status === 'in-progress'
                          ? 'info'
                          : 'neutral'
                      }
                    >
                      {selectedStep.status?.status || 'pending'}
                    </Badge>
                    {selectedStep.group_idx !== undefined && (
                      <Badge theme="neutral">
                        Group {selectedStep.group_idx}
                      </Badge>
                    )}
                  </div>
                </div>
                <div className="flex flex-col items-end gap-1">
                  {selectedStep.started_at && (
                    <Text variant="subtext" theme="neutral">
                      Started <Time time={selectedStep.started_at} format="relative" />
                    </Text>
                  )}
                  {selectedStep.finished_at && (
                    <Text variant="subtext" theme="neutral">
                      Finished <Time time={selectedStep.finished_at} format="relative" />
                    </Text>
                  )}
                  {selectedStep.execution_time && (
                    <Text variant="subtext" theme="neutral">
                      Duration: {(selectedStep.execution_time / 1000000000).toFixed(2)}s
                    </Text>
                  )}
                </div>
              </div>

              {selectedStep.status?.status_human_description && (
                <div className="p-4 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
                  <Text variant="label" theme="neutral" className="mb-1">
                    Status
                  </Text>
                  <Text variant="base">
                    {selectedStep.status.status_human_description}
                  </Text>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Text variant="label" theme="neutral" className="mb-1">
                    Step ID
                  </Text>
                  <ID>{selectedStep.id}</ID>
                </div>
                {selectedStep.idx !== undefined && (
                  <div>
                    <Text variant="label" theme="neutral" className="mb-1">
                      Index
                    </Text>
                    <Text variant="base">{selectedStep.idx}</Text>
                  </div>
                )}
                {selectedStep.execution_type && (
                  <div>
                    <Text variant="label" theme="neutral" className="mb-1">
                      Execution Type
                    </Text>
                    <Text variant="base">{selectedStep.execution_type}</Text>
                  </div>
                )}
                {selectedStep.retryable !== undefined && (
                  <div>
                    <Text variant="label" theme="neutral" className="mb-1">
                      Retryable
                    </Text>
                    <Badge theme={selectedStep.retryable ? 'success' : 'neutral'}>
                      {selectedStep.retryable ? 'Yes' : 'No'}
                    </Badge>
                  </div>
                )}
              </div>

              {selectedStep.install_workflow_id && (
                <div>
                  <Text variant="label" theme="neutral" className="mb-2">
                    Quick links
                  </Text>
                  <div className="flex flex-wrap gap-2">
                    <AdminDashboardLink
                      path={`/workflows/${selectedStep.install_workflow_id}`}
                      label="View in admin panel"
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        </Card>
      )}
    </PageSection>
  )
}
