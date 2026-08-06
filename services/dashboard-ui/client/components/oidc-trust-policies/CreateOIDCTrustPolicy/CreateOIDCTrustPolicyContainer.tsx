import type { ReactNode } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useVCSRepos } from '@/hooks/use-vcs-repos'
import { createCurrentOrgOIDCTrustPolicy } from '@/lib'
import type { TAPIError } from '@/types'
import {
  CreateOIDCTrustPolicyModal,
  type OIDCTrustPolicyFormInput,
} from './CreateOIDCTrustPolicy'

const CreateOIDCTrustPolicyModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const config = useConfig()
  const { repos, isLoading: isLoadingRepos, hasConnections } = useVCSRepos()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: OIDCTrustPolicyFormInput) =>
      createCurrentOrgOIDCTrustPolicy({
        body: {
          name: input.name,
          issuer_url: input.issuerUrl,
          audience: input.audience,
          role: input.role,
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
      }),
    onSuccess: (policy) => {
      queryClient.invalidateQueries({
        queryKey: ['oidc-trust-policies', org.id],
      })
      addToast(
        <Toast heading="Trust policy created" theme="success">
          <Text>
            {policy.name} can now exchange OIDC tokens for org access.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to create trust policy" theme="error">
          <Text>{err?.description || err?.error || 'Try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateOIDCTrustPolicyModal
      isPending={isPending}
      error={error}
      onSubmit={(input) => mutate(input)}
      repos={repos}
      isLoadingRepos={isLoadingRepos}
      hasVCSConnections={hasConnections}
      vcsConnectionsHref={`/${org.id}/connections/vcs`}
      githubAudience={config.apiUrl ?? ''}
      {...props}
    />
  )
}

export const CreateOIDCTrustPolicyButton = ({
  initialRepoFullName,
  initialRepoDefaultBranch,
  lockPreset,
  reservedNames,
  children,
  variant = 'primary',
  ...props
}: Omit<IButtonAsButton, 'children'> & {
  initialRepoFullName?: string
  initialRepoDefaultBranch?: string
  lockPreset?: boolean
  reservedNames?: string[]
  children?: ReactNode
}) => {
  const { addModal } = useSurfaces()
  const modal = (
    <CreateOIDCTrustPolicyModalContainer
      initialRepoFullName={initialRepoFullName}
      initialRepoDefaultBranch={initialRepoDefaultBranch}
      lockPreset={lockPreset}
      reservedNames={reservedNames}
    />
  )

  return (
    <Button variant={variant} onClick={() => addModal(modal)} {...props}>
      {children ?? (
        <>
          <Icon variant="PlusIcon" />
          Create trust policy
        </>
      )}
    </Button>
  )
}
