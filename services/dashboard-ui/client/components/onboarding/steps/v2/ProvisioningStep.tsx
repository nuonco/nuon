import { useEffect } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { completeDeployStep, getCurrentOnboarding, getWorkflow, getWorkflowSteps } from '@/lib'
import { getStepBadge } from '@/utils/workflow-utils'
import { toSentenceCase } from '@/utils/string-utils'
import type { TOnboarding, TWorkflow } from '@/types'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

function useOnboardingWorkflow(onboarding: TOnboarding | undefined, setSharedData: (key: string, val: unknown) => void) {
  const orgId = onboarding?.org_id
  const workflowId = onboarding?.workflow_id

  const { data: polledOnboarding } = useQuery({
    queryKey: ['onboarding-provision-poll'],
    queryFn: getCurrentOnboarding,
    enabled: !!orgId && !workflowId,
    refetchInterval: 2000,
  })

  useEffect(() => {
    if (polledOnboarding?.workflow_id && !workflowId) {
      setSharedData('onboarding', polledOnboarding)
    }
  }, [polledOnboarding, workflowId, setSharedData])

  const effectiveWorkflowId = workflowId ?? polledOnboarding?.workflow_id

  const { data: workflow } = useQuery({
    queryKey: ['onboarding-workflow', effectiveWorkflowId],
    queryFn: () => getWorkflow({ workflowId: effectiveWorkflowId!, orgId: orgId! }),
    enabled: !!effectiveWorkflowId && !!orgId,
    refetchInterval: (query) => {
      const wf = query.state.data as TWorkflow | undefined
      return wf?.finished ? false : 4000
    },
  })

  const { data: steps = [] } = useQuery({
    queryKey: ['onboarding-workflow-steps', effectiveWorkflowId],
    queryFn: () => getWorkflowSteps({ workflowId: effectiveWorkflowId!, orgId: orgId!, limit: 100, offset: 0 }),
    enabled: !!effectiveWorkflowId && !!orgId,
    refetchInterval: (query) => {
      return workflow?.finished ? false : 4000
    },
  })

  return { workflow, steps, workflowId: effectiveWorkflowId }
}

export const ProvisioningStep = ({
  onAdvance,
  onGoBack,
  sharedData,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id

  const { workflow, steps, workflowId } = useOnboardingWorkflow(onboarding, setSharedData)

  const isFinished = !!workflow?.finished
  const isError = workflow?.status?.status === 'error'

  const { mutate: completeDeploy, isPending: deployPending } = useMutation({
    mutationFn: () => completeDeployStep({ orgId: orgId! }),
    onSuccess: (ob) => {
      setSharedData('onboarding', ob)
      onAdvance()
    },
  })

  const visibleSteps = steps.filter((s) => s.execution_type !== 'hidden')

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          {isFinished && !isError ? (
            <Icon variant="CheckCircle" size={24} weight="fill" theme="success" />
          ) : isError ? (
            <Icon variant="XCircle" size={24} weight="fill" theme="error" />
          ) : (
            <Icon variant="Loading" size={24} />
          )}
          <Text variant="h3" weight="strong">
            {isFinished && !isError
              ? 'Your app is live'
              : isError
                ? 'Provisioning failed'
                : 'Provisioning...'}
          </Text>
        </div>

        {isError && workflow?.status?.status_human_description && (
          <Banner theme="error">{workflow.status.status_human_description}</Banner>
        )}
      </div>

      <div className="flex flex-col gap-3">
        {!workflowId && (
          <>
            <Skeleton height="52px" width="100%" />
            <Skeleton height="52px" width="100%" />
          </>
        )}

        {visibleSteps.map((step) => {
          const badgeConfig = getStepBadge(step)
          return (
            <Card key={step.id} className="px-4 py-3 flex items-center gap-3">
              <Status
                isWithoutText
                status={step.retried ? 'retried' : step.status?.status || 'pending'}
                variant="timeline"
                iconSize={16}
              />
              <Text weight="strong" className="flex-1 truncate">
                {toSentenceCase(step.name)}
              </Text>
              {badgeConfig?.children && (
                <Badge {...badgeConfig} size="sm" />
              )}
            </Card>
          )
        })}
      </div>

      <div className="flex justify-between">
        {onGoBack ? (
          <Button type="button" variant="secondary" onClick={onGoBack}>
            <Icon variant="CaretLeft" weight="bold" /> Back
          </Button>
        ) : (
          <div />
        )}
        <Button
          type="button"
          variant="primary"
          disabled={!isFinished || isError || deployPending}
          onClick={() => completeDeploy()}
        >
          {deployPending ? 'Completing...' : (nextStepTitle ?? 'Continue')} {!deployPending && <Icon variant="CaretRight" weight="bold" />}
        </Button>
      </div>
    </div>
  )
}
