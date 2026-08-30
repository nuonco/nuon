import type {
  TInstallStack,
  TInstallStackVersionWithCompositeError,
} from '@/types'
import type { IStepDetails } from '../../types'
import { AwaitStackDetailsContainer } from '../AwaitStackDetails'
import { GenerateStackDetails } from '../GenerateStackDetails'

export interface IStackStepDetails extends IStepDetails {
  stack?: TInstallStack
  isLoading: boolean
}

export const StackStepDetails = ({
  step,
  stack,
  isLoading,
}: IStackStepDetails) => {
  const isGenerateStack = step?.name === 'generate install stack'
  const version = stack?.versions?.at(0)
  const linksReady = !!version?.template_url || !!version?.contents

  const stackVersion = (
    stack?.versions?.find((v) => v?.id === step?.step_target_id) ??
    stack?.versions?.at(0)
  ) as TInstallStackVersionWithCompositeError | undefined

  return (
    <div>
      {isGenerateStack ? (
        <GenerateStackDetails stackVersion={stackVersion} />
      ) : (
        <AwaitStackDetailsContainer
          stack={stack}
          step={step}
          loading={isLoading || !linksReady}
        />
      )}
    </div>
  )
}
