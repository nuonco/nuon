import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deleteSlackOrgLink } from '@/lib'
import type { TAPIError, TSlackOrgLink } from '@/types'
import { DeleteOrgLinkModal } from './DeleteOrgLink'

const DeleteOrgLinkModalContainer = ({
  link,
  ...props
}: { link: TSlackOrgLink } & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () =>
      deleteSlackOrgLink({ orgId: org.id, linkId: link.id ?? '' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['slack-org-links', org.id] })
      queryClient.invalidateQueries({
        queryKey: ['slack-channel-subscriptions', org.id],
      })
      addToast(
        <Toast heading="Workspace unlinked" theme="success">
          <Text>The workspace will no longer receive events for this org.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to unlink workspace" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <DeleteOrgLinkModal
      teamId={link.team_id ?? ''}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeleteOrgLinkButton = ({
  link,
  ...props
}: { link: TSlackOrgLink } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteOrgLinkModalContainer link={link} />

  return (
    <Button
      variant="ghost"
      className="!text-red-800 dark:!text-red-500"
      onClick={() => addModal(modal)}
      {...props}
    >
      <Icon variant="TrashIcon" />
      Unlink
    </Button>
  )
}
