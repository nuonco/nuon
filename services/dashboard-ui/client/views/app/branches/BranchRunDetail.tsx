import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Badge } from '@/components/common/Badge'
import { Time } from '@/components/common/Time'
import { Icon } from '@/components/common/Icon'
import { Card } from '@/components/common/Card'
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

  // Fetch all runs and find the specific one, poll every 5 seconds
  const { data: runs = [], isLoading: isLoadingRuns } = useQuery({
    queryKey: ['branch-runs', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId,
        branchId,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000, // Poll every 5 seconds for live updates
  })

  const run = runs.find((r) => r.id === runId)
  const steps = run?.steps || []

  // Auto-select the first in-progress step or the first step
  useEffect(() => {
    if (steps.length > 0 && !selectedStep) {
      const inProgressStep = steps.find(
        (step) => step.status?.status === 'in-progress' || step.status?.status === 'running'
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
    <PageSection isScrollable>
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
      {/* Page Header */}
      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger">
            Workflow Run
          </Text>
          <ID>{runId}</ID>
          <div className="flex items-center gap-3 mt-2">
            <Badge
              theme={
                status === 'success' || status === 'completed'
                  ? 'success'
                  : status === 'failed' || status === 'error'
                  ? 'error'
                  : status === 'running' || status === 'in-progress'
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

      {/* Horizontal Workflow Canvas */}
      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <Text variant="h4" weight="strong">
              Workflow Progress
            </Text>
            <Text variant="subtext" theme="neutral">
              Scroll horizontally or use trackpad to navigate
            </Text>
          </div>
          
          {steps.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 gap-4">
              <div className="w-12 h-12 rounded-full border-4 border-blue-500 border-t-transparent animate-spin" />
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
                  const isInProgress = stepStatus === 'in-progress' || stepStatus === 'running'
                  const isSuccess = stepStatus === 'success' || stepStatus === 'completed'
                  const isError = stepStatus === 'error' || stepStatus === 'failed'
                  const isPending = !isInProgress && !isSuccess && !isError
                  
                  return (
                    <div key={step.id || idx} className="flex items-center gap-4">
                      {/* Step Card */}
                      <div
                        className={`flex flex-col items-center min-w-[240px] p-8 rounded-lg transition-all cursor-pointer border-2 ${
                          selectedStep?.id === step.id
                            ? 'ring-4 ring-purple-500 shadow-2xl scale-105 bg-purple-100 dark:bg-purple-900/20 border-purple-500'
                            : isInProgress
                            ? 'ring-2 ring-blue-500 shadow-xl hover:shadow-2xl bg-blue-100 dark:bg-blue-900/20 border-blue-500'
                            : isSuccess
                            ? 'shadow-lg hover:shadow-xl bg-green-100 dark:bg-green-900/20 border-green-500'
                            : isError
                            ? 'shadow-lg hover:shadow-xl bg-red-100 dark:bg-red-900/20 border-red-500'
                            : 'border-dashed border-cool-grey-300 dark:border-dark-grey-600 hover:border-solid hover:shadow-md bg-cool-grey-100 dark:bg-dark-grey-800/50'
                        }`}
                        onClick={() => setSelectedStep(step)}
                      >
                        {/* Step Icon */}
                        <div
                          className={`w-16 h-16 rounded-full flex items-center justify-center mb-4 transition-all ${
                            isInProgress
                              ? 'bg-blue-500 dark:bg-blue-600 text-white animate-pulse shadow-lg'
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
                            <Icon variant="Close" size={32} />
                          ) : (
                            <Icon variant="Clock" size={28} />
                          )}
                        </div>

                        {/* Step Info */}
                        <Text variant="h5" weight="stronger" className="text-center mb-2">
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
                            <Text variant="base" theme="neutral" className="font-mono font-semibold">
                              {(step.execution_time / 1000000000).toFixed(1)}s
                            </Text>
                          )}
                        </div>
                      </div>

                      {/* Connector Arrow */}
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

      {/* Selected Step Details */}
      {selectedStep && (
        <Card>
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <Text variant="h4" weight="strong">
                Step Details
              </Text>
              <button
                onClick={() => setSelectedStep(null)}
                className="text-cool-grey-600 dark:text-dark-grey-300 hover:text-cool-grey-900 dark:hover:text-white"
              >
                <Icon variant="Close" size={20} />
              </button>
            </div>
            
            <div className="space-y-4">
              {/* Step Header */}
              <div className="flex items-start justify-between">
                <div>
                  <Text variant="h5" weight="strong" className="mb-2">
                    {selectedStep.name || 'Unknown step'}
                  </Text>
                  <div className="flex items-center gap-3">
                    <Badge
                      theme={
                        selectedStep.status?.status === 'success' || selectedStep.status?.status === 'completed'
                          ? 'success'
                          : selectedStep.status?.status === 'error' || selectedStep.status?.status === 'failed'
                          ? 'error'
                          : selectedStep.status?.status === 'in-progress' || selectedStep.status?.status === 'running'
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

              {/* Status Description */}
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

              {/* Additional Info */}
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

              {/* Links */}
              {selectedStep.links && (
                <div>
                  <Text variant="label" theme="neutral" className="mb-2">
                    Quick Links
                  </Text>
                  <div className="flex flex-wrap gap-2">
                    {selectedStep.links.event_loop_ui && (
                      <a
                        href={selectedStep.links.event_loop_ui}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-2 px-3 py-2 bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300 rounded-md hover:bg-blue-100 dark:hover:bg-blue-900 transition-colors"
                      >
                        <Icon variant="ExternalLink" size={16} />
                        <Text variant="label">View in Temporal</Text>
                      </a>
                    )}
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