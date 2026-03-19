import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { cn } from '@/utils/classnames'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

type CloudSetupOption = 'cloud' | 'sandbox'

const MOCK_CURL_COMMAND = `curl -sSL https://install.nuon.co/runner | bash -s -- --token <YOUR_TOKEN>`

export const CloudSetupStep = ({ onAdvance, setSharedData, nextStepTitle }: IWizardStepComponentProps) => {
  const [selected, setSelected] = useState<CloudSetupOption | null>(null)

  const handleAdvance = () => {
    if (!selected) return
    setSharedData('cloudSetup', selected)
    onAdvance()
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-4">
        <Button
          type="button"
          variant="ghost"
          onClick={() => setSelected('cloud')}
          className="w-full !h-full !p-0"
        >
          <div
            className={cn(
              'flex flex-col w-full gap-3 p-5 border rounded-md text-left',
              selected === 'cloud' && '!bg-code/10 !border-primary-600'
            )}
          >
            <div className="flex items-center gap-3">
              <Icon variant="CloudArrowUp" size={22} />
              <Text variant="label">Connect a cloud account</Text>
            </div>
            <Text variant="body" theme="neutral" className="whitespace-normal">
              Connect your own AWS, Azure, or GCP account to deploy your application directly to your infrastructure.
            </Text>
            {selected === 'cloud' && (
              <div className="flex flex-col gap-2 mt-1 w-full">
                <Text variant="label" theme="neutral">Run this command to install the runner:</Text>
                <div className="relative w-full">
                  <CodeBlock language="bash">{MOCK_CURL_COMMAND}</CodeBlock>
                  <div className="absolute top-2 right-2">
                    <ClickToCopyButton textToCopy={MOCK_CURL_COMMAND} />
                  </div>
                </div>
              </div>
            )}
          </div>
        </Button>

        <Button
          type="button"
          variant="ghost"
          onClick={() => setSelected('sandbox')}
          className="w-full !h-full !p-0"
        >
          <div
            className={cn(
              'flex flex-col w-full gap-3 p-5 border rounded-md text-left',
              selected === 'sandbox' && '!bg-code !border-primary-600'
            )}
          >
            <div className="flex items-center gap-3">
              <Icon variant="TestTube" size={22} />
              <Text variant="label">Use sandbox mode</Text>
            </div>
            <Text variant="body" theme="neutral" className="whitespace-normal">
              We'll spin up a managed sandbox environment — no cloud account needed.
            </Text>
          </div>
        </Button>
      </div>

      <div className="flex self-end">
        <Button type="button" variant="primary" disabled={!selected} onClick={handleAdvance}>
          {nextStepTitle ?? 'Continue'} <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
