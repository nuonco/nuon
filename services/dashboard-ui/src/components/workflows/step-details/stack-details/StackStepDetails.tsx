'use client'

import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
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
    initData: {},
    path: `/api/orgs/${org.id}/installs/${step.owner_id}/stack`,
  })

  return (
    <div>
      {isGenerateStack ? (
        isLoading ? (
          <GenerateStackDetailsSkeleton />
        ) : (
          <GenerateStackDetails />
        )
      ) : isLoading ? (
        <AwaitStackDetailsSkeleton />
      ) : (
        <AwaitStackDetails stack={stack} step={step} />
      )}
    </div>
  )
}
