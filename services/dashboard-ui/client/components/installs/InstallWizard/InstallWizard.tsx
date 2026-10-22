import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { WizardStepper, type TInstallWizardStepStatus } from './WizardStepper'

export interface IInstallWizard {
  steps: { id: string; label: string; status: TInstallWizardStepStatus }[]
  selectedId: string
  title: string
  description?: string
  onSelect: (id: string) => void
  children: ReactNode
}

export const InstallWizard = ({
  steps,
  selectedId,
  title,
  description,
  onSelect,
  children,
}: IInstallWizard) => (
  <div className="flex flex-col gap-6">
    <WizardStepper steps={steps} selectedId={selectedId} onSelect={onSelect} />
    <Card elevation="1">
      <div className="flex flex-col gap-1">
        <Text variant="h3" role="heading" level={2}>
          {title}
        </Text>
        {description ? <Text theme="neutral">{description}</Text> : null}
      </div>
      {children}
    </Card>
  </div>
)
