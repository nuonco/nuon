import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deleteCurrentOrgOIDCTrustPolicy } from '@/lib'
import type { TAPIError, TOIDCTrustPolicy } from '@/types'
import { DeleteOIDCTrustPolicyModal } from './DeleteOIDCTrustPolicy'

const DeleteOIDCTrustPolicyModalContainer = ({
  policy,
  ...props
}: { policy: TOIDCTrustPolicy } & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () =>
      deleteCurrentOrgOIDCTrustPolicy({
        orgId: org.id,
        policyId: policy.id ?? '',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['oidc-trust-policies', org.id],
      })
      addToast(
        <Toast heading="Trust policy deleted" theme="success">
          <Text>{policy.name} can no longer exchange OIDC tokens.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Trust policy deletion failed" theme="error">
          <Text>{err?.description || err?.error || 'Try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <DeleteOIDCTrustPolicyModal
      policyName={policy.name ?? ''}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeleteOIDCTrustPolicyButton = ({
  policy,
  ...props
}: { policy: TOIDCTrustPolicy } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteOIDCTrustPolicyModalContainer policy={policy} />

  return (
    <Button
      variant="ghost"
      className="!text-red-800 dark:!text-red-500"
      onClick={() => addModal(modal)}
      {...props}
    >
      <Icon variant="TrashIcon" />
      Delete
    </Button>
  )
}
