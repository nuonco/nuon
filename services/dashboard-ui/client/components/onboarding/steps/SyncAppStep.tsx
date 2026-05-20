import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

export const SyncAppStep = ({ onAdvance, nextStepTitle }: IWizardStepComponentProps) => {
  const { isStepComplete } = useOnboardingJourney()
  const appSynced = isStepComplete('app_synced')

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3">
        <div className="relative">
          <ClickToCopyButton className="w-fit !absolute right-2 top-3" textToCopy="nuon apps sync" />
          <CodeBlock language="bash">nuon apps sync</CodeBlock>
        </div>
      </div>

      <Text variant="subtext" theme="neutral">
        Builds will run in the background — this usually takes a few minutes.
      </Text>

      {!appSynced && (
        <Text variant="subtext" theme="neutral">
          Waiting for app sync... Once you run <code>nuon apps sync</code>, this page will update automatically.
        </Text>
      )}

      <div className="flex justify-end">
        <Button variant="primary" disabled={!appSynced} onClick={onAdvance}>
          {nextStepTitle ?? 'Continue'} <Icon variant="CaretRightIcon" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
