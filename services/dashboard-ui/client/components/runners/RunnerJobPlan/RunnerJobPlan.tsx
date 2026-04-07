import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TRunnerJobPlan, TAPIError } from '@/types'

interface IRunnerJobPlanModal extends IModal {
  plan?: TRunnerJobPlan
  isLoading: boolean
  error: TAPIError | null
  headingText?: string
}

export const RunnerJobPlanModal = ({
  plan,
  isLoading,
  error,
  headingText = 'Runner job plan',
  ...props
}: IRunnerJobPlanModal) => {
  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
        >
          <Icon variant="CodeBlock" size="24" />
          {headingText}
        </Text>
      }
      className="!max-w-5xl !max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
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
            <Skeleton height="350px" width="100%" />
          </div>
        ) : plan ? (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <ClickToCopyButton
                textToCopy={JSON.stringify(plan, null, 2)}
                className="w-fit"
              />
            </div>
            <JSONViewer data={plan} />
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

interface IRunnerJobPlanButton extends IButtonAsButton {
  buttonText?: string
  onOpenModal: () => void
}

export const RunnerJobPlanButton = ({
  buttonText = 'View job plan',
  onOpenModal,
  ...props
}: IRunnerJobPlanButton) => {
  return (
    <Button
      onClick={() => onOpenModal()}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="CodeBlock" />}
      {buttonText}
      {props?.isMenuButton ? <Icon variant="CodeBlock" /> : null}
    </Button>
  )
}
