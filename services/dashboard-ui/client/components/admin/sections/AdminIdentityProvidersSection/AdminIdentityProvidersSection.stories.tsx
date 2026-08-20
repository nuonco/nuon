export default {
  title: 'Admin/AdminIdentityProvidersSection',
}

import type { TIdentityProvider } from '@/lib'
import { AdminIdentityProvidersSection } from './AdminIdentityProvidersSection'

const envProvider: TIdentityProvider = {
  id: 'default-google',
  provider_type: 'google',
  client_id: '822035598858-example.apps.googleusercontent.com',
  enabled: true,
  source: 'env',
}

const providers: TIdentityProvider[] = [
  envProvider,
  {
    id: 'idp0bh1hyq9woty9d7thm5afsc',
    provider_type: 'oidc',
    name: 'Staff SSO',
    client_id: 'staff-client',
    enabled: true,
    source: 'database',
  },
  {
    id: 'idpl08pd14q7g6th7dhtbgsj3t',
    provider_type: 'oidc',
    name: 'Contractor SSO',
    client_id: 'contractor-client',
    enabled: false,
    allow_all_users: false,
    source: 'database',
  },
]

export const Default = () => (
  <AdminIdentityProvidersSection
    identityProviders={providers}
    isLoading={false}
    onToggle={() => {}}
  />
)

export const Loading = () => (
  <AdminIdentityProvidersSection
    identityProviders={[]}
    isLoading
    onToggle={() => {}}
  />
)

export const Empty = () => (
  <AdminIdentityProvidersSection
    identityProviders={[]}
    isLoading={false}
    onToggle={() => {}}
  />
)

// the env provider has no database row to patch, so its toggle stays disabled
export const OnlyEnvProvider = () => (
  <AdminIdentityProvidersSection
    identityProviders={[envProvider]}
    isLoading={false}
    onToggle={() => {}}
  />
)

export const Saving = () => (
  <AdminIdentityProvidersSection
    identityProviders={providers}
    isLoading={false}
    pendingId="idp0bh1hyq9woty9d7thm5afsc"
    onToggle={() => {}}
  />
)
