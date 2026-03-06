import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { cn } from '@/utils/classnames'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

interface IExampleApp {
  id: string
  name: string
  description: string
  repo: string
  dir: string
}

const EXAMPLE_APPS: IExampleApp[] = [
  {
    id: 'eks-simple',
    name: 'EKS Simple',
    description: 'A simple Kubernetes app deployed to EKS',
    repo: 'https://github.com/nuonco/example-eks-simple',
    dir: 'example-eks-simple',
  },
  {
    id: 'aws-lambda',
    name: 'AWS Lambda',
    description: 'Serverless function deployed to AWS Lambda',
    repo: 'https://github.com/nuonco/example-aws-lambda',
    dir: 'example-aws-lambda',
  },
  {
    id: 'aws-ec2-httpbin',
    name: 'AWS EC2 / httpbin',
    description: 'HTTP testing service running on EC2',
    repo: 'https://github.com/nuonco/example-ec2-httpbin',
    dir: 'example-ec2-httpbin',
  },
  {
    id: 'coder',
    name: 'Coder',
    description: 'Remote development environments with Coder',
    repo: 'https://github.com/nuonco/example-coder',
    dir: 'example-coder',
  },
  {
    id: 'mattermost',
    name: 'Mattermost',
    description: 'Open-source team messaging platform',
    repo: 'https://github.com/nuonco/example-mattermost',
    dir: 'example-mattermost',
  },
]

export const CreateAppStep = ({ onAdvance, setSharedData, nextStepTitle }: IWizardStepComponentProps) => {
  const [selectedApp, setSelectedApp] = useState<IExampleApp | null>(null)
  const { isStepComplete } = useOnboardingJourney()
  const appCreated = isStepComplete('app_created')

  const handleSelect = (app: IExampleApp) => {
    setSelectedApp(app)
    setSharedData('selectedApp', app)
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-3">
        <div className="grid grid-cols-2 gap-3">
          {EXAMPLE_APPS.map((app) => (
            <button
              key={app.id}
              type="button"
              onClick={() => handleSelect(app)}
              className="text-left"
            >
              <Card
                className={cn(
                  'cursor-pointer transition-colors',
                  selectedApp?.id === app.id
                    ? 'border-primary-500 dark:border-primary-400 bg-primary-50 dark:bg-[#1B1026]'
                    : 'hover:border-primary-300 dark:hover:border-primary-600'
                )}
              >
                <div className="flex items-center justify-between gap-4">
                  <div className="flex flex-col gap-1">
                    <Text variant="body" weight="strong">{app.name}</Text>
                    <Text variant="subtext" theme="neutral">{app.description}</Text>
                  </div>
                  {selectedApp?.id === app.id && (
                    <div className="w-4 h-4 rounded-full bg-primary-500 flex-shrink-0" />
                  )}
                </div>
              </Card>
            </button>
          ))}
        </div>
      </div>

      {selectedApp && (
        <div className="flex flex-col gap-4">
          <Text variant="body" weight="strong">
            Set up {selectedApp.name}
          </Text>
          <div className="flex flex-col gap-3">
            <Text variant="subtext" theme="neutral">Clone the repository</Text>
            <div className="relative">
              <ClickToCopyButton className="w-fit !absolute right-2 top-3" textToCopy={`git clone ${selectedApp.repo}`} />
              <CodeBlock language="bash">{`git clone ${selectedApp.repo}`}</CodeBlock>
            </div>
          </div>
          <div className="flex flex-col gap-3">
            <Text variant="subtext" theme="neutral">Navigate to the directory</Text>
            <div className="relative">
              <ClickToCopyButton className="w-fit !absolute right-2 top-3" textToCopy={`cd ${selectedApp.dir}`} />
              <CodeBlock language="bash">{`cd ${selectedApp.dir}`}</CodeBlock>
            </div>
          </div>
          <div className="flex flex-col gap-3">
            <Text variant="subtext" theme="neutral">Create your app</Text>
            <div className="relative">
              <ClickToCopyButton className="w-fit !absolute right-2 top-3" textToCopy={`nuon apps create -n ${selectedApp.id}`} />
              <CodeBlock language="bash">{`nuon apps create -n ${selectedApp.id}`}</CodeBlock>
            </div>
          </div>
        </div>
      )}

      {!appCreated && (
        <Text variant="subtext" theme="neutral">
          Waiting for app creation... Once you run <code>nuon apps create</code>, this page will update automatically.
        </Text>
      )}

      <div className="flex justify-end">
        <Button variant="primary" disabled={!appCreated} onClick={onAdvance}>
          {nextStepTitle ?? 'Continue'} <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
