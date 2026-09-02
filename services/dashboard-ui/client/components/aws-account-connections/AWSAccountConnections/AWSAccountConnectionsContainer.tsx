import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import {
  ConnectAWSAccountButton,
  ConnectAWSAccountModalContainer,
} from '@/components/aws-account-connections/ConnectAWSAccount'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAWSAccountConnections, verifyAWSAccountConnection } from '@/lib'
import type { TAWSAccountConnection } from '@/types'
import { AWSAccountConnections } from './AWSAccountConnections'

export const AWSAccountConnectionsContainer = () => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const queryClient = useQueryClient()
  const { data = [] } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['aws-account-connections', org?.id],
    queryFn: () => getAWSAccountConnections({ orgId: org!.id }),
    enabled: !!org?.id,
  })
  const verify = useMutation({
    mutationFn: verifyAWSAccountConnection,
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: ['aws-account-connections', org?.id],
      }),
  })

  return (
    <div className="flex flex-col gap-2">
      <div className="flex justify-between items-center">
        <Text variant="subtext" weight="strong">
          AWS account connections
        </Text>
        <ConnectAWSAccountButton
          className="flex items-center gap-2 w-fit !border-transparent !p-2 !pl-1"
          size="sm"
        />
      </div>
      <AWSAccountConnections
        connections={data}
        verifyingId={
          verify.isPending ? verify.variables?.connectionId : undefined
        }
        onVerify={(connection: TAWSAccountConnection) =>
          verify.mutate({ connectionId: connection.id, orgId: org!.id })
        }
        onSelect={(connection) => {
          const modal = (
            <ConnectAWSAccountModalContainer connectionId={connection.id} />
          )
          addModal(modal)
        }}
      />
    </div>
  )
}
