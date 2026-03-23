import { useState, useEffect } from 'react'
import { Button } from '@/components/common/Button'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponentType } from '@/types/ctl-api.types'
import { cn } from '@/utils/classnames'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

const DEPLOY_PHASES = ['Queue', 'Plan', 'Apply', 'Health check', 'Live'] as const

const MOCK_COMPONENTS: { name: string; type?: TComponentType; isAction?: boolean }[] = [
  { name: 'helm-release', type: 'helm_chart' },
  { name: 'terraform-infra', type: 'terraform_module' },
  { name: 'run-scripts', isAction: true },
  { name: 'docker-workload', type: 'docker_build' },
]

type Phase = 'runner' | 'install' | 'components'

const DeployProgress = ({ currentStep }: { currentStep: number }) => (
  <div className="flex items-center w-full gap-0">
    {DEPLOY_PHASES.map((phase, i) => {
      const completed = i < currentStep
      const active = i === currentStep
      return (
        <div key={phase} className="flex items-center flex-1 last:flex-none">
          <div className="flex flex-col items-center gap-1">
            <div className="w-5 h-5 flex items-center justify-center">
              {completed ? (
                <Icon variant="CheckCircle" size={20} weight="fill" theme="success" />
              ) : active ? (
                <Icon variant="Loading" size={20} />
              ) : (
                <div className="w-4 h-4 rounded-full border-2 border-cool-grey-300 dark:border-dark-grey-600" />
              )}
            </div>
            <Text
              variant="subtext"
              theme={completed ? 'success' : active ? 'default' : 'neutral'}
              className="whitespace-nowrap text-center"
            >
              {phase}
            </Text>
          </div>
          {i < DEPLOY_PHASES.length - 1 && (
            <div
              className={cn(
                'flex-1 h-0.5 mx-1 mt-[-18px]',
                completed ? 'bg-green-500' : 'bg-cool-grey-300 dark:bg-dark-grey-600',
              )}
            />
          )}
        </div>
      )
    })}
  </div>
)

const ComponentDeployCard = ({
  name,
  type,
  isAction,
  currentStep,
}: {
  name: string
  type?: TComponentType
  isAction?: boolean
  currentStep: number
}) => {
  const isLive = currentStep >= DEPLOY_PHASES.length - 1
  return (
    <Card className="p-4 flex flex-col gap-4">
      <div className="flex items-center gap-2">
        {isLive ? (
          <Icon variant="CheckCircle" size={20} weight="fill" theme="success" />
        ) : (
          <Icon variant="Loading" size={20} />
        )}
        <Text weight="strong">{name}</Text>
        {isAction ? (
          <Icon variant="Terminal" size={14} theme="brand" />
        ) : type ? (
          <ComponentType type={type} colorVariant="color" displayVariant="icon-only" />
        ) : null}
        <div className="ml-auto">
          <Badge theme={isLive ? 'success' : 'neutral'} size="sm">
            {isLive ? 'live' : DEPLOY_PHASES[currentStep]?.toLowerCase()}
          </Badge>
        </div>
      </div>
      <DeployProgress currentStep={currentStep} />
    </Card>
  )
}

export const ProvisioningStep = ({ onAdvance, onGoBack, nextStepTitle }: IWizardStepComponentProps) => {
  const [phase, setPhase] = useState<Phase>('runner')
  const [runnerReady, setRunnerReady] = useState(false)
  const [installReady, setInstallReady] = useState(false)
  const [componentSteps, setComponentSteps] = useState<number[]>([0, -1, -1, -1])

  useEffect(() => {
    const t1 = setTimeout(() => setRunnerReady(true), 1500)
    const t2 = setTimeout(() => {
      setInstallReady(true)
      setPhase('install')
    }, 3000)
    const t3 = setTimeout(() => setPhase('components'), 4000)
    return () => {
      clearTimeout(t1)
      clearTimeout(t2)
      clearTimeout(t3)
    }
  }, [])

  useEffect(() => {
    if (phase !== 'components') return

    setComponentSteps([0, -1, -1, -1])

    const timers: ReturnType<typeof setTimeout>[] = []

    MOCK_COMPONENTS.forEach((_, compIdx) => {
      for (let step = 0; step <= DEPLOY_PHASES.length - 1; step++) {
        const delay = compIdx * 1000 + step * 600
        timers.push(
          setTimeout(() => {
            setComponentSteps((prev) => {
              const next = [...prev]
              next[compIdx] = step
              return next
            })
          }, delay),
        )
      }
    })

    return () => timers.forEach(clearTimeout)
  }, [phase])

  const allDone =
    phase === 'components' && componentSteps.every((s) => s >= DEPLOY_PHASES.length - 1)

  return (
    <div className="flex flex-col gap-6">
      {/* Section 1: Install status */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          {installReady ? (
            <Icon variant="CheckCircle" size={24} weight="fill" theme="success" />
          ) : (
            <Icon variant="Loading" size={24} />
          )}
          <Text variant="h3" weight="strong">
            {installReady ? 'Install is live' : 'Provisioning install...'}
          </Text>
        </div>
        {installReady && (
          <Text variant="body" theme="neutral">
            sandbox mode &middot; beacon-software
          </Text>
        )}

        <Card className="p-4 flex flex-col gap-3">
          <div className="flex items-center gap-3">
            <div className="w-5 h-5 flex items-center justify-center">
              {runnerReady ? (
                <Icon variant="CheckCircle" size={20} weight="fill" theme="success" />
              ) : (
                <Icon variant="Loading" size={20} />
              )}
            </div>
            <div className="flex flex-col flex-1">
              <Text weight="strong">runner-sandbox-01</Text>
              <Text variant="subtext" theme="neutral">
                {runnerReady ? 'connected \u00b7 healthy' : 'initializing...'}
              </Text>
            </div>
            <Badge theme={runnerReady ? 'success' : 'neutral'} size="sm">
              {runnerReady ? 'online' : 'pending'}
            </Badge>
          </div>

          <div className="flex items-center gap-3">
            <div className="w-5 h-5 flex items-center justify-center">
              {installReady ? (
                <Icon variant="CheckCircle" size={20} weight="fill" theme="success" />
              ) : runnerReady ? (
                <Icon variant="Loading" size={20} />
              ) : (
                <div className="w-4 h-4 rounded-full border-2 border-cool-grey-300 dark:border-dark-grey-600" />
              )}
            </div>
            <div className="flex flex-col flex-1">
              <Text weight="strong">install-abc123</Text>
              <Text variant="subtext" theme="neutral">
                {installReady ? 'healthy \u00b7 all checks passed' : 'waiting for runner...'}
              </Text>
            </div>
            <Badge theme={installReady ? 'success' : 'neutral'} size="sm">
              {installReady ? 'healthy' : 'pending'}
            </Badge>
          </div>
        </Card>
      </div>

      {/* Section 2: Component deployments */}
      {phase === 'components' && (
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-2">
            {allDone ? (
              <Icon variant="CheckCircle" size={24} weight="fill" theme="success" />
            ) : (
              <Icon variant="Loading" size={24} />
            )}
            <Text variant="h3" weight="strong">
              {allDone ? 'Your app is live' : 'Deploying components...'}
            </Text>
          </div>
          {allDone && (
            <Text variant="body" theme="neutral">
              4 components deployed &middot; install-abc123
            </Text>
          )}

          <div className="flex flex-col gap-3">
            {MOCK_COMPONENTS.map((comp, i) =>
              componentSteps[i] >= 0 ? (
                <ComponentDeployCard
                  key={comp.name}
                  name={comp.name}
                  type={comp.type}
                  isAction={comp.isAction}
                  currentStep={componentSteps[i]}
                />
              ) : null,
            )}
          </div>
        </div>
      )}

      <div className="flex justify-between">
        {onGoBack ? (
          <Button type="button" variant="secondary" onClick={onGoBack}>
            <Icon variant="CaretLeft" weight="bold" /> Back
          </Button>
        ) : (
          <div />
        )}
        <Button type="button" variant="primary" disabled={!allDone} onClick={onAdvance}>
          {nextStepTitle ?? 'Continue'} <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
