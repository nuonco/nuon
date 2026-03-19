import { useState } from 'react'
import { Tabs } from '@/components/common/Tabs'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { CloudPlatform as CloudPlatformDisplay } from '@/components/common/CloudPlatform'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

type CloudPlatform = 'aws' | 'gcp' | 'azure'
type AppAttribute =
  | 'terraform'
  | 'helm'
  | 'kubernetes'
  | 'lambda'
  | 'docker'
  | 'scripts'

const CLOUD_PLATFORMS: {
  id: CloudPlatform
  label: string
  description: string
}[] = [
  { id: 'aws', label: 'AWS', description: 'EC2, EKS, Lambda, RDS' },
  { id: 'gcp', label: 'GCP', description: 'GKE, Cloud Run, BigQuery' },
  { id: 'azure', label: 'Azure', description: 'AKS, Functions, Cosmos DB' },
]

const APP_ATTRIBUTES: {
  id: AppAttribute
  label: string
  description: string
  icon: React.ReactNode
}[] = [
  { id: 'terraform', label: 'Terraform', description: 'IaC modules', icon: <Icon variant="Terraform" size={20} /> },
  { id: 'helm', label: 'Helm charts', description: 'K8s packaging', icon: <Icon variant="Helm" size={20} /> },
  { id: 'kubernetes', label: 'Kubernetes', description: 'Raw manifests', icon: <Icon variant="Kubernetes" size={20} /> },
  { id: 'docker', label: 'Docker image', description: 'Containerized app', icon: <Icon variant="Docker" size={20} /> },
  { id: 'scripts', label: 'Custom scripts', description: 'Bash/Python', icon: <Icon variant="TerminalWindowIcon" size={20} /> },
]

interface ICustomAppTabProps {
  cloudPlatform: CloudPlatform | null
  setCloudPlatform: (p: CloudPlatform) => void
  appAttributes: AppAttribute[]
  toggleAttribute: (a: AppAttribute) => void
}

const CustomAppTab = ({
  cloudPlatform,
  setCloudPlatform,
  appAttributes,
  toggleAttribute,
}: ICustomAppTabProps) => (
  <div className="flex flex-col gap-8 pt-4">
    <div className="flex flex-col gap-1">
      <Text weight="strong">
        Create your app
      </Text>
      <Text theme="neutral">
        Tell us about your stack, or start from a real example app.
      </Text>
    </div>

    <div className="flex flex-col gap-3">
      <Text weight="strong">
        Where do your customers deploy?
      </Text>
      <div className="grid grid-cols-3 gap-3">
        {CLOUD_PLATFORMS.map((platform) => {
          const selected = cloudPlatform === platform.id
          return (
            <Button
              key={platform.id}
              type="button"
              variant="ghost"
              onClick={() => setCloudPlatform(platform.id)}
              className="w-full !h-full !p-0"
            >
              <div
                className={cn(
                  'flex w-full justify-start items-start gap-4 p-4 border rounded-md',
                  selected && '!bg-code/10 !border-primary-600'
                )}
              >
                <div
                  className={cn(
                    'mt-1.5 w-4 h-4 rounded-full border-2 shrink-0 flex items-center justify-center',
                    selected && 'border-primary-600'
                  )}
                >
                  {selected && (
                    <div className="w-2 h-2 rounded-full bg-primary-600" />
                  )}
                </div>
                <div className="flex-1 flex flex-col min-w-0 text-left">
                  <Text weight="strong">
                    {platform.label}
                  </Text>
                  <Text variant="label" theme="neutral">
                    {platform.description}
                  </Text>
                </div>
                <CloudPlatformDisplay
                  platform={platform.id}
                  colorVariant="color"
                  displayVariant="icon-only"
                  iconSize="36"
                />
              </div>
            </Button>
          )
        })}
      </div>
    </div>

    <div className="flex flex-col gap-3">
      <div className="flex items-baseline gap-2">
        <Text weight="strong">What are your app attributes?</Text>
        <Text variant="subtext" theme="neutral">
          Select all that apply
        </Text>
      </div>
      <div className="grid grid-cols-3 gap-x-6 gap-y-2">
        {APP_ATTRIBUTES.map((attr) => {
          const selected = appAttributes.includes(attr.id)
          return (
            <Button
              key={attr.id}
              type="button"
              variant="ghost"
              onClick={() => toggleAttribute(attr.id)}
              className="w-full !h-full !p-0"
            >
              <div
                className={cn(
                  'flex w-full justify-start items-start gap-4 p-4 border rounded-md',
                  selected && '!bg-code/10 !border-primary-600'
                )}
              >
                <div
                  className={cn(
                    'mt-1.5 w-4 h-4 rounded border-2 shrink-0 mt-0.5 flex items-center justify-center',
                    selected && 'bg-primary-600 border-primary-600'
                  )}
                >
                  {selected && (
                    <Icon
                      variant="Check"
                      size={10}
                      weight="bold"
                      className="text-white"
                    />
                  )}
                </div>
                <div className="flex-1 text-left flex flex-col">
                  <Text variant="body">{attr.label}</Text>
                  <Text variant="label" theme="neutral">
                    {attr.description}
                  </Text>
                </div>
                {attr.icon}
              </div>
            </Button>
          )
        })}
      </div>
    </div>
  </div>
)

const ExampleAppsTab = () => (
  <div className="flex flex-col items-center justify-center py-12 gap-4">
    <Icon variant="SquaresFour" size={40} />
    <Text theme="neutral" className="text-center">
      Example apps coming soon — check back shortly.
    </Text>
  </div>
)

export const AppProfileStep = ({
  onAdvance,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const [cloudPlatform, setCloudPlatform] = useState<CloudPlatform | null>(null)
  const [appAttributes, setAppAttributes] = useState<AppAttribute[]>([])

  const toggleAttribute = (attr: AppAttribute) => {
    setAppAttributes((prev) =>
      prev.includes(attr) ? prev.filter((a) => a !== attr) : [...prev, attr]
    )
  }

  const handleAdvance = () => {
    setSharedData('cloudPlatform', cloudPlatform)
    setSharedData('appAttributes', appAttributes)
    onAdvance()
  }

  return (
    <div className="flex flex-col gap-6">
      <Tabs
        tabs={{
          'create your own app': (
            <CustomAppTab
              cloudPlatform={cloudPlatform}
              setCloudPlatform={setCloudPlatform}
              appAttributes={appAttributes}
              toggleAttribute={toggleAttribute}
            />
          ),
          'demo using a sample app': <ExampleAppsTab />,
        }}
      />
      <div className="flex self-end">
        <Button type="button" variant="primary" onClick={handleAdvance}>
          {nextStepTitle ?? 'Continue'}{' '}
          <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
