import { useMemo, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Tabs } from '@/components/common/Tabs'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { CloudPlatform as CloudPlatformDisplay } from '@/components/common/CloudPlatform'
import { ComponentType } from '@/components/components/ComponentType'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { getExampleApps, completeYourStackStep } from '@/lib'
import { useOnboardingPoll } from '@/hooks/use-onboarding-poll'
import type { TAPIError, TExampleApp, TOnboarding, TCloudPlatform as TCloudPlatformType } from '@/types'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TComponentType } from '@/types'

const TAG_TO_COMPONENT_TYPE: Record<string, TComponentType> = {
  terraform: 'terraform_module',
  helm: 'helm_chart',
  kubernetes: 'kubernetes_manifest',
  docker: 'docker_build',
  lambda: 'job',
}

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

type SampleAppCategory = string

interface ISampleAppCardProps {
  app: TExampleApp
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
      <div className="flex items-center gap-2">
        {app.cloud_provider && (
          <CloudPlatformDisplay
            platform={app.cloud_provider as TCloudPlatformType}
            colorVariant="color"
            displayVariant="icon-only"
            iconSize="24"
          />
        )}
        {app.tags
          ?.filter((tag) => TAG_TO_COMPONENT_TYPE[tag])
          .map((tag) => (
            <ComponentType
              key={tag}
              type={TAG_TO_COMPONENT_TYPE[tag]}
              colorVariant="color"
              displayVariant="icon-only"
              iconSize="20"
            />
          ))}
      </div>
      <div className="flex flex-col gap-1">
        <Text weight="strong">{app.display_name}</Text>
        <Text variant="label" theme="neutral" className="line-clamp-2">
          {app.description}
        </Text>
      </div>
      {app.tags?.some((tag) => !TAG_TO_COMPONENT_TYPE[tag]) && (
        <div className="flex items-center gap-1.5 flex-wrap">
          {app.tags
            .filter((tag) => !TAG_TO_COMPONENT_TYPE[tag])
            .map((tag) => (
              <Badge key={tag} size="sm" variant="code">{tag}</Badge>
            ))}
        </div>
      )}
    </div>
  </Button>
)

interface IExampleAppsTabProps {
  apps: TExampleApp[]
  isLoading: boolean
  selectedApp: string | null
  onSelectApp: (slug: string) => void
  search: string
  onSearchChange: (val: string) => void
}

const CATEGORY_LABELS: Record<string, string> = {
  architecture: 'Architectures',
  'self-hosted': 'Popular self hosted applications',
}

const ExampleAppsTab = ({
  apps,
  isLoading,
  selectedApp,
  onSelectApp,
  search,
  onSearchChange,
}: IExampleAppsTabProps) => {
  const filtered = useMemo(() => {
    if (!search) return apps
    const q = search.toLowerCase()
    return apps.filter((app) =>
      (app.display_name ?? '').toLowerCase().includes(q)
    )
  }, [search, apps])

  const categories = useMemo(() => {
    const cats = new Set(filtered.map((a) => a.category).filter(Boolean))
    return Array.from(cats) as string[]
  }, [filtered])

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

      {isLoading && (
        <div className="flex items-center justify-center py-8">
          <Icon variant="Loading" size={24} />
        </div>
      )}

      {categories.map((category) => {
        const categoryApps = filtered.filter((app) => app.category === category)
        if (categoryApps.length === 0) return null
        return (
          <div className="flex flex-col gap-3" key={category}>
            <Text weight="strong">
              {CATEGORY_LABELS[category] ?? category}
            </Text>
            <div className="grid grid-cols-3 gap-3">
              {categoryApps.map((app) => (
                <SampleAppCard
                  key={app.slug}
                  app={app}
                  selected={selectedApp === app.slug}
                  onSelect={() => onSelectApp(app.slug!)}
                />
              ))}
            </div>
          </div>
        )
      })}

      {!isLoading && filtered.length === 0 && (
        <div className="flex flex-col items-center justify-center py-8 gap-2">
          <Text theme="neutral">No examples match your search.</Text>
        </div>
      )}
    </div>
  )
}

export const AppProfileStepContainer = ({
  onAdvance,
  onGoBack,
  sharedData,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const [cloudPlatform, setCloudPlatform] = useState<CloudPlatform | null>(null)
  const [appAttributes, setAppAttributes] = useState<AppAttribute[]>([])
  const [selectedSampleApp, setSelectedSampleApp] = useState<string | null>(
    null
  )
  const [sampleSearch, setSampleSearch] = useState('')
  const [waiting, setWaiting] = useState(false)

  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id

  const { data: exampleApps = [], isLoading: appsLoading } = useQuery({
    queryKey: ['onboarding-example-apps'],
    queryFn: getExampleApps,
  })

  const toggleAttribute = (attr: AppAttribute) => {
    setAppAttributes((prev) =>
      prev.includes(attr) ? prev.filter((a) => a !== attr) : [...prev, attr]
    )
  }

  const { mutate: submit, isPending, error } = useMutation({
    mutationFn: () => {
      if (!orgId) throw new Error('Organization not created yet')
      if (selectedSampleApp) {
        return completeYourStackStep({
          body: { app_type: 'example', example_app_slug: selectedSampleApp },
          orgId,
        })
      }
      return completeYourStackStep({
        body: {
          app_type: 'custom',
          cloud_provider: cloudPlatform ?? undefined,
          app_attributes: appAttributes,
        },
        orgId,
      })
    },
    onSuccess: (ob) => {
      setSharedData('onboarding', ob)
      if (ob.status_v2?.status === 'in-progress') {
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
      if (ob.status_v2?.status === 'error') return
      onAdvance()
    },
  })

  const canAdvance = selectedSampleApp || cloudPlatform
  const isWorking = isPending || waiting

  return (
    <div className="flex flex-col gap-6">
      {error && (
        <Banner theme="error">
          {(error as TAPIError).error ?? 'Failed to save app profile'}
        </Banner>
      )}
      {onboarding?.status_v2?.status === 'error' && onboarding?.step_error && (
        <Banner theme="error">{onboarding.step_error}</Banner>
      )}
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
              apps={exampleApps}
              isLoading={appsLoading}
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
        <Button
          type="button"
          variant="primary"
          disabled={!canAdvance || isWorking}
          onClick={() => submit()}
        >
          {waiting ? 'Setting up app...' : isPending ? 'Saving...' : 'Continue'}{' '}
          {!isWorking && <Icon variant="CaretRight" weight="bold" />}
        </Button>
      </div>
    </div>
  )
}
