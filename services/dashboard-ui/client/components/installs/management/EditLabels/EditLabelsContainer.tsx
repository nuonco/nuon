import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { addInstallLabels, removeInstallLabels } from '@/lib'
import { EditLabelsModal } from './EditLabels'

export const EditLabelsModalContainer = ({ ...props }: IModal) => {
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const currentLabels: Record<string, string> = install?.labels || {}

  const { mutate, isPending, error } = useMutation({
    mutationFn: async (newLabels: Record<string, string>) => {
      const currentKeys = Object.keys(currentLabels)
      const newKeys = Object.keys(newLabels)
      const removedKeys = currentKeys.filter((k) => !newKeys.includes(k))

      if (removedKeys.length > 0) {
        await removeInstallLabels({
          body: { keys: removedKeys },
          installId: install.id,
          orgId: org.id,
        })
      }

      if (newKeys.length > 0) {
        return addInstallLabels({
          body: { labels: newLabels },
          installId: install.id,
          orgId: org.id,
        })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['install', install.id] })
      addToast(
        <Toast heading="Labels updated" theme="success">
          <Text>Labels updated for {install.name}.</Text>
        </Toast>,
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Failed to update labels" theme="error">
          <Text>Unable to update labels for {install.name}.</Text>
        </Toast>,
      )
    },
  })

  return (
    <EditLabelsModal
      {...props}
      labels={currentLabels}
      isPending={isPending}
      error={error}
      onSubmit={(labels) => mutate(labels)}
    />
  )
}

export const EditLabelsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      onClick={() => {
        const modal = <EditLabelsModalContainer />
        addModal(modal)
      }}
      {...props}
    >
      Edit labels
      <Icon variant="TagIcon" />
    </Button>
  )
}
