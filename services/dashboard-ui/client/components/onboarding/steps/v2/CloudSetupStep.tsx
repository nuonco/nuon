import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useOnboardingPoll } from '@/hooks/use-onboarding-poll'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CloudPlatform as CloudPlatformDisplay } from '@/components/common/CloudPlatform'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { cn } from '@/utils/classnames'
import { completeInstallStep } from '@/lib'
import type { TAPIError, TOnboarding } from '@/types'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

type CloudSetupOption = 'cloud' | 'sandbox'
type CloudPlatform = 'aws' | 'gcp' | 'azure'

const CLOUD_LABELS: Record<CloudPlatform, string> = {
  aws: 'AWS',
  gcp: 'GCP',
  azure: 'Azure',
}

const MOCK_CURL_COMMAND = `curl -sSL https://install.nuon.co/runner | bash -s -- --token <YOUR_TOKEN>`

export const CloudSetupStep = ({
  onAdvance,
  onGoBack,
  sharedData,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const [selected, setSelected] = useState<CloudSetupOption | null>(null)
  const [waiting, setWaiting] = useState(false)
  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id
  const cloudPlatform = onboarding?.cloud_provider as CloudPlatform | null
  const cloudLabel = cloudPlatform ? CLOUD_LABELS[cloudPlatform] : null

  const { mutate: submit, isPending, error } = useMutation({
    mutationFn: () => {
      if (!orgId || !selected) throw new Error('Missing required data')
      return completeInstallStep({
        body: {
          name: onboarding?.example_app_slug
            ? `${onboarding.example_app_slug}-demo`
            : `${cloudPlatform ?? 'nuon'}-demo`,
          install_mode: selected,
        },
        orgId,
      })
    },
    onSuccess: (ob) => {
      setSharedData('onboarding', ob)
      if (ob.step_status === 'processing') {
        setWaiting(true)
      } else {
        onAdvance()
      }
    },
  })

  useOnboardingPoll({
    enabled: waiting,
    onResolved: (ob) => {
      setWaiting(false)
      setSharedData('onboarding', ob)
      if (ob.step_error) return
      onAdvance()
    },
  })

  const isWorking = isPending || waiting

  const handleAdvance = () => {
    if (!selected || isWorking) return
    submit()
  }

  return (
    <div className="flex flex-col gap-6">
      {error && (
        <Banner theme="error">
          {(error as TAPIError).error ?? 'Failed to create install'}
        </Banner>
      )}
      {onboarding?.step_error && (
        <Banner theme="error">{onboarding.step_error}</Banner>
      )}
      <div className="flex flex-col gap-12">
        <Button
          type="button"
          variant="ghost"
          onClick={() => setSelected('cloud')}
          className="w-full !h-full !p-0 focus:!bg-transparent"
        >
          <div
            className={cn(
              'flex flex-col w-full gap-3 p-5 border rounded-md text-left',
              selected === 'cloud' && '!border-primary-600'
            )}
          >
            <div className="flex items-center gap-4">
              {cloudLabel ? (
                <CloudPlatformDisplay
                  platform={cloudPlatform!}
                  colorVariant="color"
                  displayVariant="icon-only"
                  iconSize="36"
                />
              ) : (
                <Icon variant="CloudArrowUp" size="24" />
              )}
              <Text variant="base">
                Connect{' '}
                {cloudLabel ? `your ${cloudLabel} account` : 'a cloud account'}
              </Text>
            </div>
            <Text variant="body" theme="neutral" className="whitespace-normal">
              {cloudLabel
                ? `Connect your ${cloudLabel} account to deploy your application directly to your infrastructure.`
                : 'Connect your own AWS, Azure, or GCP account to deploy your application directly to your infrastructure.'}
            </Text>
            {selected === 'cloud' && (
              <div className="flex flex-col gap-2 mt-1 w-full">
                <Text variant="label" theme="neutral">
                  Run this command to install the runner:
                </Text>
                <div className="relative w-full">
                  <CodeBlock language="bash">{MOCK_CURL_COMMAND}</CodeBlock>
                  <div className="absolute top-3 right-1">
                    <ClickToCopyButton
                      className="bg-background"
                      textToCopy={MOCK_CURL_COMMAND}
                    />
                  </div>
                </div>
              </div>
            )}
          </div>
        </Button>

        <Divider dividerWord="Or" />

        <Button
          type="button"
          variant="ghost"
          onClick={() => setSelected('sandbox')}
          className="w-full !h-full !p-0 focus:!bg-transparent"
        >
          <div
            className={cn(
              'flex flex-col w-full gap-3 p-5 border rounded-md text-left',
              selected === 'sandbox' && '!border-primary-600'
            )}
          >
            <div className="flex items-center gap-4">
              <Icon variant="TestTube" size="24" />
              <Text variant="base">Use demo mode</Text>
              <Badge size="sm" theme="brand">
                Recommended
              </Badge>
            </div>
            <Text variant="body" theme="neutral" className="whitespace-normal">
              We'll spin up a managed demo environment — no cloud account
              needed.
            </Text>
          </div>
        </Button>
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
          disabled={!selected || isWorking}
          onClick={handleAdvance}
        >
          {waiting ? 'Setting up install...' : isPending ? 'Creating...' : (nextStepTitle ?? 'Continue')}{' '}
          {!isWorking && <Icon variant="CaretRight" weight="bold" />}
        </Button>
      </div>
    </div>
  )
}
