export default {
  title: 'Features / Installs / Teardown wizard',
}

import { useState, type ReactNode } from 'react'
import { InstallWizard } from '@/components/installs/InstallWizard'
import { ConfirmStep } from './ConfirmStep'
import { ReviewStep } from './ReviewStep'
import { TEARDOWN_WIZARD_STEPS, teardownWizardStatuses } from './wizard-path'

const Wizard = ({
  stepIndex,
  children,
}: {
  stepIndex: number
  children: ReactNode
}) => {
  const step = TEARDOWN_WIZARD_STEPS[stepIndex]
  return (
    <div className="bg-background p-6">
      <InstallWizard
        steps={TEARDOWN_WIZARD_STEPS.map((item, index) => ({
          id: item.id,
          label: item.navLabel,
          status: teardownWizardStatuses(stepIndex)[index],
        }))}
        selectedId={step.id}
        title={step.title}
        description={step.description}
        onSelect={() => {}}
      >
        {children}
      </InstallWizard>
    </div>
  )
}

export const Review = () => {
  const [continued, setContinued] = useState(false)
  return (
    <Wizard stepIndex={continued ? 1 : 0}>
      {continued ? (
        <ConfirmStep
          installName="acme-production"
          installHref="#"
          isPending={false}
          onGoBack={() => setContinued(false)}
          onSubmit={() => {}}
        />
      ) : (
        <ReviewStep
          installName="acme-production"
          installHref="#"
          onAdvance={() => setContinued(true)}
        />
      )}
    </Wizard>
  )
}

export const Confirm = () => (
  <Wizard stepIndex={1}>
    <ConfirmStep
      installName="acme-production"
      installHref="#"
      isPending={false}
      onGoBack={() => {}}
      onSubmit={() => {}}
    />
  </Wizard>
)
