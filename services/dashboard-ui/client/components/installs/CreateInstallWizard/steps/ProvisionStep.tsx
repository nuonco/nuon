import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { WorkflowSteps } from '@/components/workflows/WorkflowSteps'
import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { getStatusTheme } from '@/utils/status-utils'
import { isTerminalWorkflow } from '../workflow-terminal'

interface IProvisionStep {
  installName: string
  workflowId?: string
  installHref: string
}

const ProvisionProgress = ({
  installName,
  installHref,
}: {
  installName: string
  installHref: string
}) => {
  const { workflow } = useWorkflow()
  const navigate = useNavigate()
  const workflowStatus = workflow.status?.status ?? ''
  const theme = getStatusTheme(workflowStatus)
  const finished = isTerminalWorkflow({
    finished: workflow.finished,
    status: workflowStatus,
  })
  const succeeded =
    theme === 'success' ||
    (!!workflow.finished && theme !== 'error' && theme !== 'warn')

  useEffect(() => {
    if (finished) navigate(installHref, { replace: true })
  }, [finished, installHref, navigate])

  return (
    <div className="flex flex-col gap-6">
      {finished ? (
        <Banner theme={succeeded ? 'success' : theme === 'warn' ? 'warn' : 'error'}>
          {succeeded
            ? `${installName} is provisioned.`
            : `Provisioning ${installName} ${theme === 'warn' ? 'finished with warnings' : 'failed'}.`}
        </Banner>
      ) : (
        <Text theme="neutral" flex className="gap-2">
          <Icon variant="Loading" />
          Provisioning {installName}
        </Text>
      )}
      <WorkflowSteps />
      <div className="flex flex-wrap justify-end gap-3">
        {finished && succeeded ? (
          <Button variant="primary" href={installHref}>
            View install <Icon variant="CaretRightIcon" weight="bold" />
          </Button>
        ) : (
          <Button variant="secondary" href={installHref}>
            Go to install
          </Button>
        )}
      </div>
    </div>
  )
}

export const ProvisionStep = ({
  installName,
  workflowId,
  installHref,
}: IProvisionStep) => {
  if (!workflowId) {
    return (
      <div className="flex flex-col gap-6">
        <Banner theme="success">{installName} was created.</Banner>
        <div className="flex justify-end">
          <Button variant="primary" href={installHref}>
            View install <Icon variant="CaretRightIcon" weight="bold" />
          </Button>
        </div>
      </div>
    )
  }

  return (
    <WorkflowProvider workflowId={workflowId} shouldPoll>
      <ProvisionProgress installName={installName} installHref={installHref} />
    </WorkflowProvider>
  )
}
