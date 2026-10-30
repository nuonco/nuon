import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import {
  INSTALL_SCHEDULED_RECONCILES_QUERY_KEY,
  usePendingInstallScheduledReconcile,
} from '@/hooks/use-install-scheduled-reconciles'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { cancelInstallScheduledReconcile } from '@/lib'
import type { IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TInstallScheduledReconcile } from '@/types'
import { CancelScheduledReconcileModal } from './CancelScheduledReconcile'
import { PendingScheduledReconcile } from './PendingScheduledReconcile'

const CancelScheduledReconcileModalContainer = ({
  pending,
  ...props
}: { pending: TInstallScheduledReconcile } & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending, error } = useMutation<unknown, TAPIError>({
    mutationFn: () =>
      cancelInstallScheduledReconcile({
        orgId: org.id,
        installId: install.id,
        scheduledReconcileId: pending.id!,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Scheduled reconcile canceled" theme="success">
          <Text>The scheduled reconcile for {install.name} was canceled.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({
        queryKey: [INSTALL_SCHEDULED_RECONCILES_QUERY_KEY],
      })
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Unable to cancel schedule" theme="error">
          <Text>
            Unable to cancel the scheduled reconcile for {install.name}.
          </Text>
        </Toast>
      )
    },
  })

  return (
    <CancelScheduledReconcileModal
      pending={pending}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const PendingScheduledReconcileContainer = ({
  size,
}: {
  size?: 'sm' | 'md'
}) => {
  const { addModal } = useSurfaces()
  const { pending } = usePendingInstallScheduledReconcile()

  if (!pending) return null

  return (
    <PendingScheduledReconcile
      pending={pending}
      size={size}
      onCancel={() =>
        addModal(<CancelScheduledReconcileModalContainer pending={pending} />)
      }
    />
  )
}
