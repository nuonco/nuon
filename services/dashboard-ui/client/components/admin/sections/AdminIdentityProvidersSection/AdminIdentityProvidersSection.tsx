import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { Toggle } from '@/components/common/form/Toggle'
import type { TIdentityProvider } from '@/lib'
import { AdminSection } from '../../shared/AdminSection'

export interface IAdminIdentityProvidersSection {
  identityProviders: TIdentityProvider[]
  isLoading: boolean
  pendingId?: string
  onToggle: (identityProvider: TIdentityProvider, enabled: boolean) => void
}

export const AdminIdentityProvidersSection = ({
  identityProviders,
  isLoading,
  pendingId,
  onToggle,
}: IAdminIdentityProvidersSection) => (
  <AdminSection
    title="Identity providers"
    subtitle="Sign-in providers for this deployment. Disabling one removes its button from the sign-in page; existing sessions are unaffected."
  >
    {isLoading && <Text variant="subtext">Loading identity providers…</Text>}

    {!isLoading && identityProviders.length === 0 && (
      <Text variant="subtext">No identity providers configured.</Text>
    )}

    <div className="space-y-3">
      {identityProviders.map((identityProvider) => (
        <div
          key={identityProvider.id}
          className="flex items-center justify-between gap-4 p-4 rounded-lg border border-gray-200 dark:border-gray-700"
        >
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-2">
              <Text variant="base" weight="strong">
                {identityProvider.name || identityProvider.provider_type}
              </Text>
              <Badge size="xs">{identityProvider.provider_type}</Badge>
              {identityProvider.source === 'env' && (
                <Badge size="xs">env</Badge>
              )}
            </div>
            <Text
              variant="subtext"
              className="text-gray-600 dark:text-gray-300"
            >
              {identityProvider.client_id}
            </Text>
          </div>

          <Toggle
            checked={identityProvider.enabled}
            // the env provider has no database row to patch, so it can only be turned off by
            // removing its config from the environment
            disabled={
              identityProvider.source === 'env' ||
              pendingId === identityProvider.id
            }
            onChange={(enabled) => onToggle(identityProvider, enabled)}
            label={`Enable ${identityProvider.name || identityProvider.provider_type}`}
          />
        </div>
      ))}
    </div>
  </AdminSection>
)
