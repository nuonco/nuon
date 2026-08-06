import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CreateOIDCTrustPolicyButton } from '@/components/oidc-trust-policies/CreateOIDCTrustPolicy'
import { DeleteOIDCTrustPolicyButton } from '@/components/oidc-trust-policies/DeleteOIDCTrustPolicy'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  getCurrentOrgOIDCTrustPolicies,
  updateCurrentOrgOIDCTrustPolicy,
} from '@/lib'
import type { TAPIError, TOIDCTrustPolicy } from '@/types'
import { ManageRepoOIDCModal } from './ManageRepoOIDC'

const ManageRepoOIDCModalContainer = ({
  defaultBranch,
  repoFullName,
  ...props
}: { defaultBranch: string; repoFullName: string } & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { addToast } = useToast()

  const {
    data: policies,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['oidc-trust-policies', org.id],
    queryFn: () => getCurrentOrgOIDCTrustPolicies({ orgId: org.id }),
  })

  const repoPolicies = (policies ?? []).filter((policy) =>
    policy.claim_conditions?.sub?.startsWith(`repo:${repoFullName}:`)
  )

  const {
    mutate: toggle,
    isPending: isToggling,
    variables: togglingVars,
  } = useMutation({
    mutationFn: ({
      policy,
      next,
    }: {
      policy: TOIDCTrustPolicy
      next: boolean
    }) =>
      updateCurrentOrgOIDCTrustPolicy({
        body: { enabled: next },
        orgId: org.id,
        policyId: policy.id ?? '',
      }),
    onSuccess: (_res, { policy, next }) => {
      queryClient.invalidateQueries({
        queryKey: ['oidc-trust-policies', org.id],
      })
      addToast(
        <Toast
          heading={next ? 'Trust policy enabled' : 'Trust policy disabled'}
          theme="success"
        >
          <Text>
            {policy.name} {next ? 'can now' : 'can no longer'} exchange OIDC
            tokens.
          </Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to update trust policy" theme="error">
          <Text>{err?.description || err?.error || 'Try again.'}</Text>
        </Toast>
      )
    },
  })

  const reservedNames = (policies ?? []).map((policy) => policy.name ?? '')

  return (
    <ManageRepoOIDCModal
      policies={repoPolicies}
      isLoading={isLoading}
      error={error as TAPIError | null}
      togglingId={isToggling ? togglingVars?.policy.id : undefined}
      onToggle={(policy, next) => toggle({ policy, next })}
      renderDelete={(policy) => (
        <DeleteOIDCTrustPolicyButton policy={policy} size="sm" />
      )}
      createSlot={
        <CreateOIDCTrustPolicyButton
          variant="secondary"
          size="sm"
          className="w-fit"
          initialRepoFullName={repoFullName}
          initialRepoDefaultBranch={defaultBranch}
          reservedNames={reservedNames}
          lockPreset
        >
          <Icon variant="PlusIcon" size={14} />
          Create trust policy
        </CreateOIDCTrustPolicyButton>
      }
      {...props}
    />
  )
}

export const ManageRepoOIDCButton = ({
  defaultBranch,
  repoFullName,
  ...props
}: { defaultBranch: string; repoFullName: string } & Omit<
  IButtonAsButton,
  'children'
>) => {
  const { addModal } = useSurfaces()
  const modal = (
    <ManageRepoOIDCModalContainer
      defaultBranch={defaultBranch}
      repoFullName={repoFullName}
    />
  )

  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="ShieldCheckIcon" size={14} />
      Manage OIDC
    </Button>
  )
}
