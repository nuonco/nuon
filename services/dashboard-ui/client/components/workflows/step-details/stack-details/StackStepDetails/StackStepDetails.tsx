import type { TInstallStack } from '@/types'
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

  return (
    <div>
      {isGenerateStack ? (
        <GenerateStackDetails />
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
