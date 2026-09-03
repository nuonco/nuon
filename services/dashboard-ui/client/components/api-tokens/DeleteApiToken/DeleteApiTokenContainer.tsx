import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deleteStaticToken } from '@/lib'
import type { TStaticToken } from '@/types'
import { DeleteApiTokenModal } from './DeleteApiToken'

const tokenLabel = (token: TStaticToken) => token.name || 'This token'

const DeleteApiTokenModalContainer = ({
  token,
  ...props
}: { token: TStaticToken } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () => deleteStaticToken({ tokenId: token.id, orgId: org.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['static-tokens', org?.id] })
      addToast(
        <Toast heading="Token deleted" theme="success">
          <Text>
            {tokenLabel(token)} was deleted and can no longer access the API.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Delete failed" theme="error">
          <Text>There was an error deleting {tokenLabel(token)}.</Text>
        </Toast>
      )
    },
  })

  return (
    <DeleteApiTokenModal
      tokenName={tokenLabel(token)}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeleteApiTokenButton = ({
  token,
  ...props
}: { token: TStaticToken } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteApiTokenModalContainer token={token} />

  return (
    <Button
      variant="ghost"
      className="!text-red-800 dark:!text-red-500 !p-2 w-full justify-between"
      onClick={() => addModal(modal)}
      {...props}
    >
      Delete token
      {props?.isMenuButton ? <Icon variant="TrashIcon" /> : null}
    </Button>
  )
}
