import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'

interface IReviewStep {
  installName: string
  installHref: string
  onAdvance: () => void
}

export const ReviewStep = ({
  installName,
  installHref,
  onAdvance,
}: IReviewStep) => (
  <div className="flex flex-col gap-8">
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <Text weight="strong">What will happen</Text>
        <Text theme="neutral">
          Deprovisioning {installName} will remove it from the cloud account.
        </Text>
      </div>
      <div className="flex flex-col gap-3">
        <Text>This will create a workflow that attempts to:</Text>
        <ul className="flex list-disc flex-col gap-1 pl-6">
          <li>
            <Text theme="neutral">
              Teardown each install component according to the dependency order.
            </Text>
          </li>
          <li>
            <Text theme="neutral">Teardown the install sandbox.</Text>
          </li>
        </ul>
      </div>
      <Banner theme="warn">
        <Text>
          <strong>Important:</strong> After this workflow completes, manually
          teardown the install stack.
        </Text>
      </Banner>
    </div>
    <div className="flex flex-wrap items-center justify-end gap-3">
      <Button type="button" variant="secondary" href={installHref}>
        Cancel teardown
      </Button>
      <Button type="button" variant="primary" onClick={onAdvance}>
        Continue <Icon variant="CaretRightIcon" weight="bold" />
      </Button>
    </div>
  </div>
)
