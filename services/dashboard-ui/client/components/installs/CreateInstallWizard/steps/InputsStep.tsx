import { useStore } from '@tanstack/react-form'
import { Text } from '@/components/common/Text'
import { InstallInputFields } from '@/components/installs/forms/InstallForm/InstallInputFields'
import type { InstallFormApi } from '@/components/installs/forms/InstallForm'
import type { TAppInputConfig } from '@/types'
import { installInputsBlocker } from '../readiness'
import { StepActions } from '../StepActions'

interface IInputsStep {
  form: InstallFormApi
  inputConfig?: TAppInputConfig
  onAdvance: () => void
  onGoBack?: () => void
  cancelHref?: string
}

export const InputsStep = ({
  form,
  inputConfig,
  onAdvance,
  onGoBack,
  cancelHref,
}: IInputsStep) => {
  const values = useStore(form.store, (state) => state.values)
  const blocker = installInputsBlocker(values, inputConfig)

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-3">
        <Text variant="subtext" theme="neutral">
          Inputs configure the selected app.
        </Text>
        <InstallInputFields form={form} inputConfig={inputConfig} />
      </div>
      <StepActions
        onGoBack={onGoBack}
        onAdvance={onAdvance}
        label="Continue"
        showCaret
        cancelHref={cancelHref}
        disabled={!!blocker}
        disabledReason={blocker}
      />
    </div>
  )
}
