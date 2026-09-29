import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
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
    repo: 'https://github.com/nuonco/example-app-configs',
    dir: 'example-app-configs/eks-simple',
  },
  {
    id: 'cde',
    name: 'AWS EC2 and Claude Code',
    description: 'A VM accessible with SSH and VS Code Web',
    repo: 'https://github.com/nuonco/example-app-configs',
    dir: 'example-app-configs/cde',
  },
  {
    id: 'aws-lambda',
    name: 'AWS Lambda',
    description: 'Serverless function deployed to AWS Lambda',
    repo: 'https://github.com/nuonco/example-app-configs',
    dir: 'example-app-configs/aws-lambda',
  },
  {
    id: 'httpbin',
    name: 'AWS EC2 Simple',
    description: 'HTTP testing service running on EC2',
    repo: 'https://github.com/nuonco/example-app-configs',
    dir: 'example-app-configs/httpbin',
  },
  {
    id: 'ecs-simple',
    name: 'AWS ECS Simple',
    description: 'A simple containerized service deployed to AWS ECS',
    repo: 'https://github.com/nuonco/example-app-configs',
    dir: 'example-app-configs/ecs-simple',
  },
  {
    id: 'coder',
    name: 'Coder & Grafana on AWS EKS',
    description: 'Dev environments for Claude Code agents and devs with coder.com',
    repo: 'https://github.com/nuonco/example-app-configs',
    dir: 'example-app-configs/coder',
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
        <fieldset className="grid grid-cols-2 gap-2">
          {EXAMPLE_APPS.map((app) => {
            const isSelected = selectedApp?.id === app.id
            return (
              <label
                key={app.id}
                htmlFor={app.id}
                className={cn(
                  'flex items-center gap-3 px-4 py-3 rounded-md border shadow-sm cursor-pointer transition-colors',
                  isSelected
                    ? '!border-primary-600 dark:!border-primary-400 bg-primary-50 dark:bg-primary-950/40'
                    : 'hover:!border-primary-300 dark:hover:!border-primary-700'
                )}
              >
                <input
                  type="radio"
                  id={app.id}
                  name="example-app"
                  value={app.id}
                  checked={isSelected}
                  onChange={() => handleSelect(app)}
                  className="accent-primary-600 flex-shrink-0"
                />
                <div className="flex flex-col gap-0.5 min-w-0">
                  <Text variant="body" weight="strong">{app.name}</Text>
                  <Text variant="subtext" theme="neutral" as="div" className="truncate">{app.description}</Text>
                </div>
              </label>
            )
          })}
        </fieldset>
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
            <Text variant="subtext" theme="neutral">Create your app (required to proceed)</Text>
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
          {nextStepTitle ?? 'Continue'} <Icon variant="CaretRightIcon" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
