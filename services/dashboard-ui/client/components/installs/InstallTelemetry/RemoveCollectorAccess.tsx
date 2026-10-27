import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteInstallTelemetryCollector } from '@/lib/ctl-api/installs/install-telemetry-collector'
import { deleteToken } from '@/lib/ctl-api/tokens/delete-token'
import type { TToken } from '@/types'

export const RemoveCollectorAccessModal = ({
  installId,
  orgId,
  token,
  ...props
}: {
  installId: string
  orgId: string
  token?: TToken
} & IModal) => {
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const mutation = useMutation({
    mutationFn: () =>
      token
        ? deleteToken({ orgId, tokenId: token.id! })
        : deleteInstallTelemetryCollector({ orgId, installId }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['install-telemetry-collector', orgId, installId],
      })
      queryClient.invalidateQueries({ queryKey: ['tokens', orgId] })
      queryClient.invalidateQueries({ queryKey: ['service-accounts', orgId] })
      removeModal(props.modalId)
      addToast(
        <Toast
          heading={token ? 'Credential revoked' : 'Collector access removed'}
          theme="success"
        >
          <Text>
            Already-issued relay credentials can remain valid for up to ten
            minutes.
          </Text>
        </Toast>
      )
    },
  })
  return (
    <Modal
      heading={
        token ? 'Revoke bootstrap credential?' : 'Remove collector access?'
      }
      primaryActionTrigger={{
        children: mutation.isPending
          ? 'Removing...'
          : token
            ? 'Revoke credential'
            : 'Remove collector access',
        variant: 'danger',
        disabled: mutation.isPending,
        onClick: () => mutation.mutate(),
      }}
      {...props}
    >
      <FormErrorBanner
        error={mutation.error}
        fallback="Access removal failed"
      />
      <Banner theme="warn">
        {token
          ? 'Verify that the replacement credential is in use before revoking this credential.'
          : 'This revokes every bootstrap credential and grant for this install’s collector account.'}
      </Banner>
      <Text>
        {token
          ? `Revoke ${token.name || token.id}.`
          : 'Remove the managed collector service account.'}{' '}
        This cannot be undone. Issued relay credentials can remain valid for up
        to ten minutes.
      </Text>
      <Text variant="subtext">
        This does not uninstall the collector, remove queue storage, change the
        forwarding policy, or affect the runner's identity.
      </Text>
    </Modal>
  )
}
