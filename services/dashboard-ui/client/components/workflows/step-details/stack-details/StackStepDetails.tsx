'use client'

import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import type { TInstallStack } from '@/types'
import type { IStepDetails } from '../types'
import {
  AwaitStackDetails,
  AwaitStackDetailsSkeleton,
} from './AwaitStackDetails'
import {
  GenerateStackDetails,
  GenerateStackDetailsSkeleton,
} from './GenerateStackDetails'

interface IStackStepDetails extends IStepDetails {}

export const StackStepDetails = ({ step }: IStackStepDetails) => {
  const isGenerateStack = step.name === 'generate install stack'
  const { org } = useOrg()
  const { data: stack, isLoading } = useQuery<TInstallStack>({
    queryKey: ['install-stack', org?.id, step?.owner_id],
    queryFn: () => getInstallStack({ orgId: org.id, installId: step.owner_id }),
    enabled: !!org?.id && !!step?.owner_id,
    refetchInterval: (query) => {
      if (isGenerateStack) return false
      const hasLinks = !!query.state.data?.versions?.at(0)?.template_url
      return hasLinks ? false : 3000
    },
  })

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
        <AwaitStackDetailsSkeleton />
      ) : (
        <AwaitStackDetails stack={stack} step={step} />
      )}
    </div>
  )
}
