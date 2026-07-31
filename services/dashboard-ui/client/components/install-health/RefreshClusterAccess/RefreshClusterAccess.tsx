import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { refreshInstallHealthClusterAccess } from '@/lib'
import type { TAPIError } from '@/types'

interface IRefreshClusterAccessModal extends IModal {
  installId: string
}

export const RefreshClusterAccessModal = ({
  installId,
  ...props
}: IRefreshClusterAccessModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [roleName, setRoleName] = useState<string>('')

  const { mutate: refresh, isPending } = useMutation({
    mutationFn: () =>
      refreshInstallHealthClusterAccess({
        orgId: org!.id,
        installId,
        roleName,
      }),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['install-health-timeline'] })
      queryClient.invalidateQueries({
        queryKey: ['install-component-health-timeline'],
      })

      if (!data?.cluster_found) {
        addToast(
          <Toast heading="No cluster found" theme="warn">
            <Text>
              This install&apos;s sandbox does not create a Kubernetes cluster,
              so there are no workloads for health to watch.
            </Text>
          </Toast>
        )
      } else {
        addToast(
          <Toast heading="Cluster access refreshed" theme="success">
            <Text>
              Health will read {data.cluster_id} as {data.role_name} within a
              minute.
            </Text>
          </Toast>
        )
      }
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Refresh failed" theme="error">
          <Text>{err?.error || 'Unable to refresh cluster access.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading="Refresh cluster access"
      primaryActionTrigger={{
        children: isPending ? 'Refreshing...' : 'Refresh access',
        disabled: isPending,
        onClick: () => refresh(),
      }}
      {...props}
    >
      <Text>
        Health reads this install&apos;s cluster to report component status.
        Refreshing rebuilds that access from the install&apos;s current outputs
        — useful when components show unknown because the install has not been
        deployed since health was enabled, or after the cluster endpoint changed.
      </Text>
      <Text>
        Unless you pick another role, health reads as the maintenance role — the
        same one drift scans and action runs use.
      </Text>
      <RoleSelector
        installId={installId}
        value={roleName}
        onChange={setRoleName}
        disabled={isPending}
      />
    </Modal>
  )
}

export const RefreshClusterAccessButton = ({
  installId,
}: {
  installId: string
}) => {
  const { addModal } = useSurfaces()
  const modal = <RefreshClusterAccessModal installId={installId} />

  return (
    <Button variant="ghost" size="xs" onClick={() => addModal(modal)}>
      <Icon variant="ArrowsClockwiseIcon" size={14} />
      Refresh cluster access
    </Button>
  )
}
