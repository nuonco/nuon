import { Button } from '@/components/common/Button'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

export const DownloadCliStep = ({
  onAdvance,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const { isStepComplete } = useOnboardingJourney()
  const cliInstalled = isStepComplete('cli_installed')

  return (
    <div className="flex flex-col gap-6">
      <Tabs
        tabs={{
          homebrew: (
            <div className="mt-4 relative">
              <ClickToCopyButton
                className="w-fit !absolute right-2 top-3"
                textToCopy="brew install nuonco/tap/nuon"
              />
              <CodeBlock language="bash">
                brew install nuonco/tap/nuon
              </CodeBlock>
            </div>
          ),
          script: (
            <div className="mt-4 relative">
              <ClickToCopyButton
                className="w-fit !absolute right-2 top-3"
                textToCopy={`bash <(curl -sSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/install.sh)`}
              />
              <CodeBlock language="bash">{`bash <(curl -sSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/install.sh)`}</CodeBlock>
              <Text variant="subtext" theme="neutral">
                The script automatically detects your OS and architecture,
                downloads the latest version, and installs it to your path.
              </Text>
            </div>
          ),
        }}
      />

      <div className="flex flex-col gap-2">
        <Text variant="body" weight="strong">
          Authenticate
        </Text>
        <div className="relative">
          <ClickToCopyButton
            className="w-fit !absolute right-2 top-3"
            textToCopy="nuon auth login"
          />
          <CodeBlock language="bash">nuon auth login</CodeBlock>
        </div>
      </div>

      <a
        href="https://docs.nuon.co/cli"
        target="_blank"
        rel="noopener noreferrer"
        className="text-primary-600 dark:text-primary-400 text-sm underline underline-offset-2 w-fit"
      >
        View CLI documentation
      </a>

      {!cliInstalled && (
        <Text variant="subtext" theme="neutral">
          Waiting for CLI authentication... Once you run <code>nuon auth login</code>, this page will update automatically.
        </Text>
      )}

      <div className="flex justify-end">
        <Button variant="primary" disabled={!cliInstalled} onClick={onAdvance}>
          {nextStepTitle ?? 'Continue'}{' '}
          <Icon variant="CaretRightIcon" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
