import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import type { TInstallStack } from '@/types'
import type { IStepDetails } from '../../types'
import { StackStepDetails } from './StackStepDetails'

interface IStackStepDetailsContainer extends IStepDetails {}

export const StackStepDetailsContainer = ({ step }: IStackStepDetailsContainer) => {
  const isGenerateStack = step?.name === 'generate install stack'
  const { org } = useOrg()
  const { data: stack, isLoading } = useQuery<TInstallStack>({
    queryKey: ['install-stack', org?.id, step?.owner_id],
    queryFn: () => getInstallStack({ orgId: org!.id, installId: step!.owner_id }),
    enabled: !!org?.id && !!step?.owner_id,
    refetchInterval: (query) => {
      if (isGenerateStack) return false
      const hasLinks = !!query.state.data?.versions?.at(0)?.template_url
      return hasLinks ? false : 3000
    },
  })

  return (
    <StackStepDetails
      step={step}
      stack={stack}
      isLoading={isLoading}
    />
  )
}
