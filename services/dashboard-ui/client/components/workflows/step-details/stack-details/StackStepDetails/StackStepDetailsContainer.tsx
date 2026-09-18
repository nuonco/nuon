import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import type { TInstallStack, TInstallStackVersion } from '@/types'
import type { IStepDetails } from '../../types'
import { StackStepDetails } from './StackStepDetails'

interface IStackStepDetailsContainer extends IStepDetails {}

const STEP_TERMINAL_STATUSES = new Set([
  'error',
  'success',
  'cancelled',
  'discarded',
])

export const StackStepDetailsContainer = ({
  step,
}: IStackStepDetailsContainer) => {
  const isGenerateStack = step?.name === 'generate install stack'
  const { org } = useOrg()
  const { data: stack, isLoading } = useQuery<TInstallStack>({
    queryKey: ['install-stack', org?.id, step?.owner_id],
    queryFn: () =>
      getInstallStack({ orgId: org!.id, installId: step!.owner_id }),
    enabled: !!org?.id && !!step?.owner_id,
    refetchInterval: (query) => {
      const stepStatus = step?.status?.status
      const isStepTerminal = STEP_TERMINAL_STATUSES.has(stepStatus ?? '')

      const versions = query.state.data?.versions ?? []
      const version = (versions.find((v) => v?.id === step?.step_target_id) ??
        versions.at(0)) as TInstallStackVersion | undefined
      const hasVersionData = !!(
        version?.composite_error ||
        version?.template_url ||
        version?.contents
      )

      if (isGenerateStack) {
        return isStepTerminal && hasVersionData ? false : 3000
      }

      // Keep polling past the launch links: the run carrying the stack
      // outputs only lands once the customer applies the stack, and the
      // outputs card stays empty until it does.
      const hasRun = !!version?.runs?.length
      if (isStepTerminal && (hasRun || stepStatus !== 'success')) return false

      return hasVersionData ? 10000 : 3000
    },
  })

  return <StackStepDetails step={step} stack={stack} isLoading={isLoading} />
}
