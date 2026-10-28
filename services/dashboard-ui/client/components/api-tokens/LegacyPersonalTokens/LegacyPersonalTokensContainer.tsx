import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteStaticToken, listStaticTokens } from '@/lib'
import type { TStaticToken } from '@/types'
import { LegacyPersonalTokens } from './LegacyPersonalTokens'

const RevokeLegacyTokenModal = ({
  token,
  ...props
}: { token: TStaticToken } & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () => deleteStaticToken({ tokenId: token.id, orgId: org.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['static-tokens', org?.id] })
      addToast(
        <Toast heading="Token revoked" theme="success">
          <Text>{token.name || 'This token'} can no longer access the API.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong" theme="error">
          Revoke personal token?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? 'Revoking token' : 'Revoke token',
        disabled: isPending,
        onClick: () => mutate(),
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <FormErrorBanner error={error} fallback="Unable to revoke token" />
        <Banner theme="warn">
          Consider creating a service account token to replace it first.
        </Banner>
        <Text variant="body">
          {token.name || 'This token'} will be revoked and can no longer be
          used to access the API. This cannot be undone.
        </Text>
      </div>
    </Modal>
  )
}

export const LegacyPersonalTokensContainer = () => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()

  const { data, isLoading } = useQuery({
    queryKey: ['static-tokens', org.id],
    queryFn: () => listStaticTokens({ orgId: org.id }),
  })

  const personalTokens = (data ?? []).filter(
    (token) => !!token.account_id && token.account_id === token.created_by_id
  )

  return (
    <LegacyPersonalTokens
      data={personalTokens}
      isLoading={isLoading}
      onRevoke={(token) => addModal(<RevokeLegacyTokenModal token={token} />)}
    />
  )
}
