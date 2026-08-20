import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useAuth } from '@/hooks/use-auth'
import { useToast } from '@/hooks/use-toast'
import {
  adminGetIdentityProviders,
  adminSetIdentityProviderEnabled,
  type TIdentityProvider,
} from '@/lib'
import { AdminIdentityProvidersSection } from './AdminIdentityProvidersSection'

export const AdminIdentityProvidersSectionContainer = () => {
  const { addToast } = useToast()
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const adminEmail = user?.email ?? ''

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'identity-providers'],
    queryFn: adminGetIdentityProviders,
  })

  const { mutate: setEnabled, variables } = useMutation({
    mutationFn: ({
      identityProvider,
      enabled,
    }: {
      identityProvider: TIdentityProvider
      enabled: boolean
    }) =>
      adminSetIdentityProviderEnabled({
        identityProviderId: identityProvider.id,
        enabled,
        adminEmail,
      }),
    onSuccess: (_result, { identityProvider, enabled }) => {
      queryClient.invalidateQueries({
        queryKey: ['admin', 'identity-providers'],
      })
      addToast(
        <Toast
          heading={enabled ? 'Provider enabled' : 'Provider disabled'}
          theme="success"
        >
          <Text>{identityProvider.name || identityProvider.provider_type}</Text>
        </Toast>
      )
    },
    onError: () => {
      addToast(
        <Toast heading="Update failed" theme="error">
          <Text>Unable to update the identity provider. Try again.</Text>
        </Toast>
      )
    },
  })

  return (
    <AdminIdentityProvidersSection
      identityProviders={data ?? []}
      isLoading={isLoading}
      pendingId={variables?.identityProvider.id}
      onToggle={(identityProvider, enabled) =>
        setEnabled({ identityProvider, enabled })
      }
    />
  )
}
