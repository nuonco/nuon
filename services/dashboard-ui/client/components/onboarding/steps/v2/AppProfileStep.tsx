import { useMemo, useState } from 'react'
import { Tabs } from '@/components/common/Tabs'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { CloudPlatform as CloudPlatformDisplay } from '@/components/common/CloudPlatform'
import { ComponentType } from '@/components/components/ComponentType'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TComponentType } from '@/types'

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
  {
    id: 'terraform',
    label: 'Terraform',
    description: 'IaC modules',
    icon: (
      <ComponentType
        type="terraform_module"
        colorVariant="color"
        displayVariant="icon-only"
        iconSize="24"
      />
    ),
  },
  {
    id: 'helm',
    label: 'Helm charts',
    description: 'K8s packaging',
    icon: (
      <ComponentType
        type="helm_chart"
        colorVariant="color"
        displayVariant="icon-only"
        iconSize="24"
      />
    ),
  },
  {
    id: 'kubernetes',
    label: 'Kubernetes',
    description: 'Raw manifests',
    icon: (
      <ComponentType
        type="kubernetes_manifest"
        colorVariant="color"
        displayVariant="icon-only"
        iconSize="24"
      />
    ),
  },
  {
    id: 'docker',
    label: 'Docker image',
    description: 'Containerized app',
    icon: (
      <ComponentType
        type="docker_build"
        colorVariant="color"
        displayVariant="icon-only"
        iconSize="24"
      />
    ),
  },
  {
    id: 'scripts',
    label: 'Custom scripts',
    description: 'Bash/Python',
    icon: <Icon variant="TerminalWindowIcon" size={24} theme="success" />,
  },
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
      <Text weight="strong">Create your app</Text>
      <Text theme="neutral">
        Tell us about your stack, or start from a real example app.
      </Text>
    </div>

    <div className="flex flex-col gap-3">
      <Text weight="strong">Where do your customers deploy?</Text>
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
                  <Text weight="strong">{platform.label}</Text>
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

type SampleAppCategory = 'architecture' | 'self-hosted'

interface ISampleApp {
  id: string
  name: string
  description: string
  category: SampleAppCategory
  cloud: CloudPlatform
  componentTypes: TComponentType[]
}

const SAMPLE_APPS: ISampleApp[] = [
  {
    id: 'httpbin',
    name: 'Httpbin',
    description:
      'HTTP request & response debugging service. Deploys an EC2 instance with Docker',
    category: 'architecture',
    cloud: 'aws',
    componentTypes: ['docker_build'],
  },
  {
    id: 'aws-lambda-api-gateway',
    name: 'AWS Lambda + API Gateway',
    description:
      'Lambda function from a Go Docker image with DynamoDB, API Gateway and a certificate',
    category: 'architecture',
    cloud: 'aws',
    componentTypes: ['docker_build', 'terraform_module', 'job'],
  },
  {
    id: 'eks-simple',
    name: 'EKS Simple',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'architecture',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
  {
    id: 'eks-auto-mode',
    name: 'EKS Auto mode',
    description:
      'HTTP request & response debugging service. Deploys an EC2 instance with Docker',
    category: 'architecture',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'docker_build'],
  },
  {
    id: 'gke-simple',
    name: 'GKE Simple',
    description:
      'Lambda function from a Go Docker image with DynamoDB, API Gateway and a certificate',
    category: 'architecture',
    cloud: 'gcp',
    componentTypes: ['terraform_module', 'docker_build', 'job'],
  },
  {
    id: 'grafana',
    name: 'Grafana',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'self-hosted',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
  {
    id: 'clickhouse',
    name: 'Clickhouse',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'self-hosted',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
  {
    id: 'baserow',
    name: 'Baserow',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'self-hosted',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
  {
    id: 'penpot',
    name: 'Penpot',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'self-hosted',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
  {
    id: 'coder',
    name: 'Coder',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'self-hosted',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
  {
    id: 'twenty-crm',
    name: 'Twenty CRM',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'self-hosted',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
  {
    id: 'mattermost',
    name: 'Mattermost',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'self-hosted',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
  {
    id: 'clickhouse-tailscale',
    name: 'Clickhouse + Tailscale',
    description:
      'EKS cluster with a Whoami workload, Application Load Balancer and a certificate',
    category: 'self-hosted',
    cloud: 'aws',
    componentTypes: ['terraform_module', 'helm_chart'],
  },
]

const SAMPLE_APP_CATEGORIES: {
  id: SampleAppCategory
  label: string
}[] = [
  { id: 'architecture', label: 'Architectures' },
  { id: 'self-hosted', label: 'Popular self hosted applications' },
]

interface ISampleAppCardProps {
  app: ISampleApp
  selected: boolean
  onSelect: () => void
}

const SampleAppCard = ({ app, selected, onSelect }: ISampleAppCardProps) => (
  <Button
    type="button"
    variant="ghost"
    onClick={onSelect}
    className="w-full !h-full !p-0"
  >
    <div
      className={cn(
        'flex w-full h-full flex-col gap-3 p-4 border rounded-md text-left',
        selected && '!bg-code/10 !border-primary-600'
      )}
    >
      <div className="flex flex-col gap-1">
        <Text weight="strong">{app.name}</Text>
        <Text variant="label" theme="neutral" className="line-clamp-2">
          {app.description}
        </Text>
      </div>
      <div className="flex items-center gap-1.5">
        <CloudPlatformDisplay
          platform={app.cloud}
          colorVariant="color"
          displayVariant="icon-only"
          iconSize="16"
        />
        {app.componentTypes.map((type) => (
          <ComponentType
            key={type}
            type={type}
            colorVariant="color"
            displayVariant="icon-only"
            iconSize="16"
          />
        ))}
      </div>
    </div>
  </Button>
)

interface IExampleAppsTabProps {
  selectedApp: string | null
  onSelectApp: (id: string) => void
  search: string
  onSearchChange: (val: string) => void
}

const ExampleAppsTab = ({
  selectedApp,
  onSelectApp,
  search,
  onSearchChange,
}: IExampleAppsTabProps) => {
  const filtered = useMemo(() => {
    if (!search) return SAMPLE_APPS
    const q = search.toLowerCase()
    return SAMPLE_APPS.filter((app) => app.name.toLowerCase().includes(q))
  }, [search])

  return (
    <div className="flex flex-col gap-6 pt-4">
      <div className="flex flex-col gap-1">
        <Text weight="strong">Start from an example</Text>
        <Text theme="neutral">
          Deploy a working app to see Nuon in action. Customize or replace it
          once you're ready.
        </Text>
      </div>

      <SearchInput
        value={search}
        onChange={onSearchChange}
        placeholder="Search examples"
      />

      {SAMPLE_APP_CATEGORIES.map((category) => {
        const apps = filtered.filter((app) => app.category === category.id)
        if (apps.length === 0) return null
        return (
          <div className="flex flex-col gap-3" key={category.id}>
            <Text weight="strong">{category.label}</Text>
            <div className="grid grid-cols-3 gap-3">
              {apps.map((app) => (
                <SampleAppCard
                  key={app.id}
                  app={app}
                  selected={selectedApp === app.id}
                  onSelect={() => onSelectApp(app.id)}
                />
              ))}
            </div>
          </div>
        )
      })}

      {filtered.length === 0 && (
        <div className="flex flex-col items-center justify-center py-8 gap-2">
          <Text theme="neutral">No examples match your search.</Text>
        </div>
      )}
    </div>
  )
}

export const AppProfileStep = ({
  onAdvance,
  onGoBack,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const [cloudPlatform, setCloudPlatform] = useState<CloudPlatform | null>(null)
  const [appAttributes, setAppAttributes] = useState<AppAttribute[]>([])
  const [selectedSampleApp, setSelectedSampleApp] = useState<string | null>(
    null
  )
  const [sampleSearch, setSampleSearch] = useState('')

  const toggleAttribute = (attr: AppAttribute) => {
    setAppAttributes((prev) =>
      prev.includes(attr) ? prev.filter((a) => a !== attr) : [...prev, attr]
    )
  }

  const handleAdvance = () => {
    const sampleApp = SAMPLE_APPS.find((a) => a.id === selectedSampleApp)
    setSharedData('cloudPlatform', sampleApp?.cloud ?? cloudPlatform)
    setSharedData('appAttributes', appAttributes)
    setSharedData('sampleApp', selectedSampleApp)
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
          'demo using a sample app': (
            <ExampleAppsTab
              selectedApp={selectedSampleApp}
              onSelectApp={setSelectedSampleApp}
              search={sampleSearch}
              onSearchChange={setSampleSearch}
            />
          ),
        }}
      />
      <div className="flex justify-between">
        {onGoBack ? (
          <Button type="button" variant="secondary" onClick={onGoBack}>
            <Icon variant="CaretLeft" weight="bold" /> Back
          </Button>
        ) : (
          <div />
        )}
        <Button type="button" variant="primary" onClick={handleAdvance}>
          {nextStepTitle ?? 'Continue'}{' '}
          <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
