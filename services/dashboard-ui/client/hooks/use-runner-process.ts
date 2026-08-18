import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunnerProcess } from '@/lib'

export const useRunnerProcess = ({
  runnerId,
  processId,
}: {
  runnerId?: string
  processId?: string
}) => {
  const { org } = useOrg()

  return useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runner-process', org?.id, runnerId, processId],
    queryFn: () =>
      getRunnerProcess({
        orgId: org!.id,
        runnerId: runnerId!,
        processId: processId!,
      }),
    enabled: !!org?.id && !!runnerId && !!processId,
  })
}
