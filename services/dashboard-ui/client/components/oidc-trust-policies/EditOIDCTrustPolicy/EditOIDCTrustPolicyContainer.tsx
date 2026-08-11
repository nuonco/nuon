import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useRoleOptions } from '@/hooks/use-roles'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateCurrentOrgOIDCTrustPolicy } from '@/lib'
import type { TOIDCTrustPolicy } from '@/types'
import {
  EditOIDCTrustPolicyModal,
  type EditOIDCTrustPolicyFormInput,
} from './EditOIDCTrustPolicy'

const EditOIDCTrustPolicyModalContainer = ({
  policy,
  ...props
}: { policy: TOIDCTrustPolicy } & Record<string, any>) => {
  const { org } = useOrg()
  const { roleOptions } = useRoleOptions('oidc_trust_policy')
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: EditOIDCTrustPolicyFormInput) =>
      updateCurrentOrgOIDCTrustPolicy({
        body: {
          name: input.name,
          issuer_url: input.issuerUrl,
          audience: input.audience,
          role: input.role,
          enabled: input.enabled,
          ...(input.tokenDurationSeconds
            ? { token_duration_seconds: Number(input.tokenDurationSeconds) }
            : {}),
          claim_conditions: Object.fromEntries(
            input.claimConditions
              .filter((condition) => condition.key.trim())
              .map((condition) => [
                condition.key.trim(),
                condition.value.trim(),
              ])
          ),
        },
        orgId: org.id,
        policyId: policy.id ?? '',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['oidc-trust-policies', org.id],
      })
      addToast(
        <Toast heading="Trust policy updated" theme="success">
          <Text>Changes to {policy.name} are live.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <EditOIDCTrustPolicyModal
      policy={policy}
      isPending={isPending}
      error={error}
      roleOptions={roleOptions}
      onSubmit={(input) => mutate(input)}
      {...props}
    />
  )
}

export const EditOIDCTrustPolicyButton = ({
  policy,
  ...props
}: { policy: TOIDCTrustPolicy } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <EditOIDCTrustPolicyModalContainer policy={policy} />

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PencilSimpleIcon" />
      Edit
    </Button>
  )
}
