import { useState } from 'react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  deleteCloudConnection,
  getCloudConnections,
  verifyCloudConnection,
} from '@/lib'
import type { TCloudConnection } from '@/types'
import { CreateCloudConnectionModal } from '../CreateCloudConnection/CreateCloudConnection'
import { CloudConnectionsTable as Table } from './CloudConnectionsTable'

const ManageCloudConnectionModalContainer = ({
  initialConnection,
  ...props
}: { initialConnection: TCloudConnection } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [connection, setConnection] = useState(initialConnection)
  const verifyMutation = useMutation({
    mutationFn: () =>
      verifyCloudConnection({ orgId: org.id, connectionId: connection.id }),
    onSuccess: (verified) => {
      setConnection(verified)
      queryClient.invalidateQueries({
        queryKey: ['cloud-connections', org.id],
      })
      const succeeded = verified.status === 'verified'
      addToast(
        <Toast
          heading={succeeded ? 'Connection verified' : 'Verification failed'}
          theme={succeeded ? 'success' : 'error'}
        >
          <Text>
            {succeeded
              ? `${verified.name} is ready to use.`
              : verified.status_message ||
                `${verified.name} could not be verified.`}
          </Text>
        </Toast>
      )
    },
  })

  return (
    <CreateCloudConnectionModal
      connection={connection}
      error={null}
      isPending={false}
      isVerifying={verifyMutation.isPending}
      verifyError={verifyMutation.error}
      onSubmit={() => {}}
      onVerify={() => verifyMutation.mutate()}
      onDone={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const CloudConnectionsTable = () => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [pendingActionId, setPendingActionId] = useState<string>()

  const query = useQuery({
    queryKey: ['cloud-connections', org.id],
    queryFn: () => getCloudConnections({ orgId: org.id }),
    placeholderData: keepPreviousData,
  })

  const verifyMutation = useMutation({
    mutationFn: (connection: TCloudConnection) => {
      setPendingActionId(connection.id)
      return verifyCloudConnection({
        orgId: org.id,
        connectionId: connection.id,
      })
    },
    onSuccess: (connection) => {
      queryClient.invalidateQueries({ queryKey: ['cloud-connections', org.id] })
      const succeeded = connection.status === 'verified'
      addToast(
        <Toast
          heading={succeeded ? 'Connection verified' : 'Verification failed'}
          theme={succeeded ? 'success' : 'error'}
        >
          <Text>
            {succeeded
              ? `${connection.name} is ready to use.`
              : connection.status_message ||
                `${connection.name} could not be verified.`}
          </Text>
        </Toast>
      )
    },
    onError: (_error, connection) => {
      addToast(
        <Toast heading="Verification failed" theme="error">
          <Text>{connection.name} could not be verified.</Text>
        </Toast>
      )
    },
    onSettled: () => setPendingActionId(undefined),
  })

  const deleteMutation = useMutation({
    mutationFn: (connection: TCloudConnection) => {
      setPendingActionId(connection.id)
      return deleteCloudConnection({
        orgId: org.id,
        connectionId: connection.id,
      })
    },
    onSuccess: (_data, connection) => {
      queryClient.invalidateQueries({ queryKey: ['cloud-connections', org.id] })
      addToast(
        <Toast heading="Connection deleted" theme="success">
          <Text>{connection.name} was deleted.</Text>
        </Toast>
      )
    },
    onSettled: () => setPendingActionId(undefined),
  })

  const openConnection = (connection: TCloudConnection) => {
    addModal(
      <ManageCloudConnectionModalContainer initialConnection={connection} />
    )
  }

  return (
    <Table
      connections={query.data ?? []}
      isLoading={query.isLoading}
      pendingActionId={pendingActionId}
      onOpen={openConnection}
      onVerify={(connection) => verifyMutation.mutate(connection)}
      onDelete={(connection) => {
        if (window.confirm(`Delete ${connection.name}?`))
          deleteMutation.mutate(connection)
      }}
    />
  )
}
