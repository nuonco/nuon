import type { TInstallStack } from '@/types'
import type { IStepDetails } from '../../types'
import {
  AwaitStackDetailsContainer,
  AwaitStackDetailsSkeletonContainer,
} from '../AwaitStackDetails'
import {
  GenerateStackDetails,
  GenerateStackDetailsSkeleton,
} from '../GenerateStackDetails'

export interface IStackStepDetails extends IStepDetails {
  stack?: TInstallStack
  isLoading: boolean
}

export const StackStepDetails = ({ step, stack, isLoading }: IStackStepDetails) => {
  const isGenerateStack = step?.name === 'generate install stack'
  const version = stack?.versions?.at(0)
  const linksReady = !!version?.template_url || !!version?.contents

  return (
    <div>
      {isGenerateStack ? (
        isLoading && !stack ? (
          <GenerateStackDetailsSkeleton />
        ) : (
          <GenerateStackDetails />
        )
      ) : isLoading || !linksReady ? (
        <AwaitStackDetailsSkeletonContainer />
      ) : (
        <AwaitStackDetailsContainer stack={stack} step={step} />
      )}
    </div>
  )
}
