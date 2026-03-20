import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CloudPlatform as CloudPlatformDisplay } from '@/components/common/CloudPlatform'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { cn } from '@/utils/classnames'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

type CloudSetupOption = 'cloud' | 'sandbox'
type CloudPlatform = 'aws' | 'gcp' | 'azure'

const CLOUD_LABELS: Record<CloudPlatform, string> = {
  aws: 'AWS',
  gcp: 'GCP',
  azure: 'Azure',
}

const MOCK_CURL_COMMAND = `curl -sSL https://install.nuon.co/runner | bash -s -- --token <YOUR_TOKEN>`

export const CloudSetupStep = ({ onAdvance, onGoBack, sharedData, setSharedData, nextStepTitle }: IWizardStepComponentProps) => {
  const [selected, setSelected] = useState<CloudSetupOption | null>(null)
  const cloudPlatform = sharedData.cloudPlatform as CloudPlatform | null
  const cloudLabel = cloudPlatform ? CLOUD_LABELS[cloudPlatform] : null

  const handleAdvance = () => {
    if (!selected) return
    setSharedData('cloudSetup', selected)
    onAdvance()
  }

  return (
    <div className="flex flex-col gap-6">
      {cloudLabel && (
        <div className="flex items-center gap-2">
          <CloudPlatformDisplay
            platform={cloudPlatform!}
            colorVariant="color"
            displayVariant="icon-only"
            iconSize="36"
          />
          <Text theme="neutral">Deploying to {cloudLabel}</Text>
        </div>
      )}

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
              <Text variant="base">
                Connect {cloudLabel ? `your ${cloudLabel} account` : 'a cloud account'}
              </Text>
            </div>
            <Text variant="body" theme="neutral" className="whitespace-normal">
              {cloudLabel
                ? `Connect your ${cloudLabel} account to deploy your application directly to your infrastructure.`
                : 'Connect your own AWS, Azure, or GCP account to deploy your application directly to your infrastructure.'}
            </Text>
            {selected === 'cloud' && (
              <div className="flex flex-col gap-2 mt-1 w-full">
                <Text variant="label" theme="neutral">Run this command to install the runner:</Text>
                <div className="relative w-full">
                  <CodeBlock language="bash">{MOCK_CURL_COMMAND}</CodeBlock>
                  <div className="absolute top-3 right-1">
                    <ClickToCopyButton className="bg-background" textToCopy={MOCK_CURL_COMMAND} />
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
              <Text variant="base">Use demo mode</Text>
            </div>
            <Text variant="body" theme="neutral" className="whitespace-normal">
              We'll spin up a managed demo environment — no cloud account needed.
            </Text>
          </div>
        </Button>
      </div>

      <div className="flex justify-between">
        {onGoBack ? (
          <Button type="button" variant="secondary" onClick={onGoBack}>
            <Icon variant="CaretLeft" weight="bold" /> Back
          </Button>
        ) : <div />}
        <Button type="button" variant="primary" disabled={!selected} onClick={handleAdvance}>
          {nextStepTitle ?? 'Continue'} <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
