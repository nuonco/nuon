import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createCloudConnection, verifyCloudConnection } from '@/lib'
import type { TCloudConnection } from '@/types'
import { CreateCloudConnectionModal } from './CreateCloudConnection'
import type { CreateCloudConnectionValues } from './schema'

const CreateCloudConnectionModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [connection, setConnection] = useState<TCloudConnection | null>(null)

  const createMutation = useMutation({
    mutationFn: (values: CreateCloudConnectionValues) => {
      return createCloudConnection({
        orgId: org.id,
        body: {
          name: values.name.trim(),
          platform: 'aws',
          target_id: values.targetId.trim(),
          principal: values.principal.trim(),
          preset: values.preset,
        },
      })
    },
    onSuccess: (created) => {
      setConnection(created)
      queryClient.invalidateQueries({ queryKey: ['cloud-connections', org.id] })
    },
  })

  const verifyMutation = useMutation({
    mutationFn: () =>
      verifyCloudConnection({
        orgId: org.id,
        connectionId: connection!.id,
      }),
    onSuccess: (verified) => {
      setConnection(verified)
      queryClient.invalidateQueries({ queryKey: ['cloud-connections', org.id] })
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
      error={createMutation.error}
      isPending={createMutation.isPending}
      isVerifying={verifyMutation.isPending}
      verifyError={verifyMutation.error}
      onSubmit={(values) => createMutation.mutate(values)}
      onVerify={() => verifyMutation.mutate()}
      onDone={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const CreateCloudConnectionButton = (
  props: Omit<IButtonAsButton, 'children'>
) => {
  const { addModal } = useSurfaces()
  return (
    <Button
      variant="secondary"
      onClick={() => addModal(<CreateCloudConnectionModalContainer />)}
      {...props}
    >
      <Icon variant="PlusIcon" />
      Create connection
    </Button>
  )
}
