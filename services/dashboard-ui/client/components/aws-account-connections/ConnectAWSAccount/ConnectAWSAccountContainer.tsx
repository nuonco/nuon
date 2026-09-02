import { useState } from 'react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  createAWSAccountConnection,
  getAWSAccountConnection,
  updateAWSAccountConnection,
  verifyAWSAccountConnection,
} from '@/lib'
import type { TAWSAccountConnection, TAPIError } from '@/types'
import { ConnectAWSAccountModal } from './ConnectAWSAccount'

export const ConnectAWSAccountModalContainer = ({
  connectionId,
  ...props
}: IModal & { connectionId?: string }) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [connection, setConnection] = useState<TAWSAccountConnection>()
  const [error, setError] = useState<TAPIError | null>(null)
  const {
    data: existingConnection,
    error: existingConnectionError,
    isPending: isLoadingExistingConnection,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['aws-account-connection', org?.id, connectionId],
    queryFn: () =>
      getAWSAccountConnection({ connectionId: connectionId!, orgId: org!.id }),
    enabled: !!org?.id && !!connectionId,
  })
  const currentConnection = connection ?? existingConnection

  const refreshConnections = () =>
    queryClient.invalidateQueries({
      queryKey: ['aws-account-connections', org?.id],
    })

  const create = useMutation({
    mutationFn: createAWSAccountConnection,
    onSuccess: (data) => {
      setError(null)
      setConnection(data)
      refreshConnections()
    },
    onError: (err: TAPIError) => setError(err),
  })
  const update = useMutation({
    mutationFn: updateAWSAccountConnection,
    onSuccess: (data) => {
      setError(null)
      setConnection(data)
      refreshConnections()
    },
    onError: (err: TAPIError) => setError(err),
  })
  const verify = useMutation({
    mutationFn: verifyAWSAccountConnection,
    onSuccess: (data) => {
      setError(null)
      setConnection(data)
      refreshConnections()
      if (data.verification_status === 'verified') {
        addToast(
          <Toast heading="AWS account connected" theme="success">
            <Text>Verified access to AWS account {data.account_id}.</Text>
          </Toast>
        )
        removeModal(props.modalId)
      }
    },
    onError: (err: TAPIError) => setError(err),
  })

  return (
    <ConnectAWSAccountModal
      connection={currentConnection}
      error={error ?? (existingConnectionError as TAPIError | null)}
      isLoading={!!connectionId && isLoadingExistingConnection}
      isVerifying={verify.isPending}
      isPending={
        (!!connectionId && isLoadingExistingConnection) ||
        create.isPending ||
        update.isPending ||
        verify.isPending
      }
      onCreate={({ name, accountId, region }) =>
        create.mutate({
          body: { name, account_id: accountId, default_region: region },
          orgId: org!.id,
        })
      }
      onSetRole={(roleArn) =>
        update.mutate({
          connectionId: currentConnection!.id,
          body: { role_arn: roleArn },
          orgId: org!.id,
        })
      }
      onVerify={() =>
        verify.mutate({ connectionId: currentConnection!.id, orgId: org!.id })
      }
      {...props}
    />
  )
}

export const ConnectAWSAccountButton = ({
  children = 'Add',
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ConnectAWSAccountModalContainer />
  return (
    <Button onClick={() => addModal(modal)} {...props}>
      <Icon variant="PlusIcon" />
      {children}
    </Button>
  )
}
