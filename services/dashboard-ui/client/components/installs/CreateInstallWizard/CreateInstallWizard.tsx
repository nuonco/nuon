import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { CreateStepper } from './CreateStepper'
import type { TCreateWizardStepStatus } from './wizard-path'

export interface ICreateInstallWizard {
  steps: { id: string; label: string; status: TCreateWizardStepStatus }[]
  selectedId: string
  title: string
  description?: string
  onSelect: (id: string) => void
  children: ReactNode
}

export const CreateInstallWizard = ({
  steps,
  selectedId,
  title,
  description,
  onSelect,
  children,
}: ICreateInstallWizard) => (
  <div className="flex flex-col gap-6">
    <CreateStepper steps={steps} selectedId={selectedId} onSelect={onSelect} />
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
