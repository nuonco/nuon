export default {
  title: 'Features / Service accounts / Service accounts table',
}

import { useState } from 'react'
import type {
  TServiceAccount,
  TServiceAccountManagement,
} from '@/types'
import { ServiceAccountFilters } from '@/components/service-accounts/ServiceAccountFilters/ServiceAccountFilters'
import { ServiceAccountsTable } from './ServiceAccountsTable'

const ORG_ID = 'org-acme'

const base = {
  created_at: '2026-09-01T00:00:00Z',
  updated_at: '2026-09-01T00:00:00Z',
  account_type: 'service',
} as const

const userManaged: TServiceAccount = {
  ...base,
  id: 'acc-ci-deploy',
  name: 'ci-deploy',
  email: 'svc-ci-deploy@example.com',
  roles: [{ id: 'role-builder', role_type: 'org_builder', title: 'Builder' }],
  system_account: false,
  purposes: [],
  managed_service_account: null,
}

const multipleRoles: TServiceAccount = {
  ...base,
  id: 'acc-release-bot',
  name: 'release-bot',
  email: 'svc-release-bot@example.com',
  roles: [
    { id: 'role-admin', role_type: 'org_admin', title: 'Admin' },
    { id: 'role-installer', role_type: 'installer', title: 'Installer' },
    { id: 'role-custom-auditor', role_type: 'org_read_only', title: 'Auditor' },
  ],
  system_account: false,
  purposes: [],
  managed_service_account: null,
}

const unnamed: TServiceAccount = {
  ...base,
  id: 'acc-unnamed',
  email: 'svc-4f2a9c@example.com',
  roles: [
    { id: 'role-read-only', role_type: 'org_read_only', title: 'Read-only' },
  ],
  system_account: false,
  purposes: [],
  managed_service_account: null,
}

const longName: TServiceAccount = {
  ...base,
  id: 'acc-long-name',
  name: 'acme-production-us-east-1-terraform-cloud-workspace-automation-bot',
  email:
    'svc-acme-production-us-east-1-terraform-cloud-workspace-automation-bot@example.com',
  roles: [{ id: 'role-builder', role_type: 'org_builder', title: 'Builder' }],
  system_account: false,
  purposes: [],
  managed_service_account: null,
}

const managedStack: TServiceAccount = {
  ...base,
  id: 'acc-stack-acme-prod',
  name: 'stack-acme-prod',
  email: 'svc-stack-acme-prod@example.com',
  roles: [{ id: 'role-private-stack-1', role_type: 'stack' }],
  system_account: true,
  purposes: ['stack'],
  managed_service_account: {
    owner_type: 'install_stacks',
    owner_id: 'stk-acme-prod',
    purpose: 'stack',
    instance_key: 'default',
    owner_name: 'acme-prod',
    install_id: 'inst-acme-prod',
  },
}

const managedTelemetryPreview: TServiceAccount = {
  ...base,
  id: 'acc-telemetry-acme-staging',
  name: 'telemetry-acme-staging',
  email: 'svc-telemetry-acme-staging@example.com',
  roles: [
    {
      id: 'role-private-telemetry-1',
      role_type: 'runner',
      title: 'Telemetry relay',
    },
  ],
  system_account: true,
  purposes: ['telemetry'],
  managed_service_account: {
    owner_type: 'installs',
    owner_id: 'inst-acme-staging',
    purpose: 'telemetry',
    instance_key: 'default',
    owner_name: 'acme-staging',
  },
}

const legacyRunner: TServiceAccount = {
  ...base,
  id: 'acc-legacy-runner',
  name: 'runner-acme-dev',
  email: 'svc-runner-acme-dev@example.com',
  roles: [{ id: 'role-private-runner-1', role_type: 'runner' }],
  system_account: true,
  purposes: ['runner'],
  managed_service_account: null,
}

const legacyStack: TServiceAccount = {
  ...base,
  id: 'acc-legacy-stack',
  email: 'svc-stack-acme-legacy@example.com',
  roles: [{ id: 'role-private-stack-2', role_type: 'stack' }],
  system_account: true,
  purposes: ['stack'],
  managed_service_account: null,
}

const userAccounts = [userManaged, multipleRoles, unnamed, longName]
const systemAccounts = [
  managedStack,
  managedTelemetryPreview,
  legacyRunner,
  legacyStack,
]

const pagination = { hasNext: false, offset: 0, limit: 20 }

export const UserManaged = () => (
  <ServiceAccountsTable
    data={userAccounts}
    orgId={ORG_ID}
    isLoading={false}
    pagination={pagination}
  />
)

export const System = () => (
  <ServiceAccountsTable
    data={systemAccounts}
    orgId={ORG_ID}
    isLoading={false}
    management="system"
    pagination={pagination}
  />
)

export const All = () => (
  <ServiceAccountsTable
    data={[...userAccounts, ...systemAccounts]}
    orgId={ORG_ID}
    isLoading={false}
    management="all"
    pagination={{ ...pagination, hasNext: true }}
  />
)

export const WithFilterToolbar = () => {
  const [management, setManagement] =
    useState<TServiceAccountManagement>('user')
  const [purpose, setPurpose] = useState('')
  const [search, setSearch] = useState('')
  const pool =
    management === 'user'
      ? userAccounts
      : management === 'system'
        ? systemAccounts
        : [...userAccounts, ...systemAccounts]
  const term = search.trim().toLowerCase()
  const data = pool.filter(
    (account) =>
      (!purpose || account.purposes?.includes(purpose)) &&
      (!term ||
        [
          account.name,
          account.email,
          account.id,
          account.managed_service_account?.owner_name,
          account.managed_service_account?.owner_id,
        ].some((value) => value?.toLowerCase().includes(term)))
  )

  return (
    <ServiceAccountsTable
      data={data}
      orgId={ORG_ID}
      isLoading={false}
      management={management}
      hasActiveFilters={!!purpose || !!term}
      pagination={pagination}
      filterActions={
        <ServiceAccountFilters
          management={management}
          purpose={purpose}
          purposes={['runner', 'stack', 'telemetry']}
          search={search}
          onManagementChange={(value) => {
            setManagement(value)
            if (value === 'user') setPurpose('')
          }}
          onPurposeChange={setPurpose}
          onSearchChange={setSearch}
        />
      }
    />
  )
}

export const Empty = () => (
  <ServiceAccountsTable
    data={[]}
    orgId={ORG_ID}
    isLoading={false}
    pagination={pagination}
  />
)

export const EmptySystem = () => (
  <ServiceAccountsTable
    data={[]}
    orgId={ORG_ID}
    isLoading={false}
    management="system"
    pagination={pagination}
  />
)

export const EmptyFiltered = () => (
  <ServiceAccountsTable
    data={[]}
    orgId={ORG_ID}
    isLoading={false}
    management="all"
    hasActiveFilters
    pagination={pagination}
  />
)

export const Loading = () => (
  <ServiceAccountsTable
    data={[]}
    orgId={ORG_ID}
    isLoading
    pagination={pagination}
  />
)

export const LoadError = () => (
  <ServiceAccountsTable
    data={[]}
    orgId={ORG_ID}
    isLoading={false}
    error="Refresh the page to try again."
    pagination={pagination}
  />
)
