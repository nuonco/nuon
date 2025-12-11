'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Skeleton } from '@/components/common/Skeleton'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useQuery } from '@/hooks/use-query'
import type { TRunnerJobPlan } from '@/types'

interface IRunnerJobPlan {
  runnerJobId: string
  buttonText?: string
  headingText?: string
}

export const RunnerJobPlanModal = ({ 
  runnerJobId,
  headingText = 'Runner job plan',
  ...props 
}: Omit<IRunnerJobPlan, "buttonText"> & IModal) => {
  const { org } = useOrg()

  const {
    data: plan,
    error,
    isLoading,
  } = useQuery<TRunnerJobPlan>({
    path: `/api/orgs/${org?.id}/runners/jobs/${runnerJobId}/plan`,
    dependencies: [runnerJobId]
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="CodeBlock" size="24" />
          {headingText}
        </Text>
      }
      className="!max-w-5xl"
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to load runner job plan.'}
          </Banner>
        ) : null}

        {isLoading ? (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <Skeleton height="26px" width="26px" />
            </div>
            <Skeleton height="600px" width="100%" />
          </div>
        ) : plan ? (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <ClickToCopyButton
                textToCopy={JSON.stringify(plan, null, 2)}
                className="w-fit"
              />
            </div>
            <CodeBlock 
              language="json"
              className="max-h-[600px]"
            >
              {JSON.stringify(plan, null, 2)}
            </CodeBlock>
          </div>
        ) : (
          <div className="flex items-center justify-center p-8">
            <Text variant="body" theme="neutral">
              No plan data available
            </Text>
          </div>
        )}
      </div>
    </Modal>
  )
}

export const RunnerJobPlanButton = ({
  runnerJobId,
  buttonText = 'View job plan',
  headingText,
  ...props
}: IRunnerJobPlan & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = (
    <RunnerJobPlanModal 
      runnerJobId={runnerJobId}
      headingText={headingText}
    />
  )

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      <Icon variant="CodeBlock" />
      {buttonText}
    </Button>
  )
}
