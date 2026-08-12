import { useState } from 'react'
import type { TPermissionEntry } from '@/types'
import { readAllWriteScoped } from '../permissions'
import { PermissionEntries } from './PermissionEntries'

export default {
  title: 'Access roles/PermissionEntries',
}

const ORG_ID = 'orgrok933tcyzji01s7us3aeo3'

const appOptions = [
  { value: 'app98e2wpzdxwoey393edtqj45', label: 'checkout' },
  { value: 'appq7fplr1up5atx5zpxotbabm', label: 'billing' },
]

const installOptions = [
  { value: 'inl4plkdhwau58atwfd92vlc8q', label: 'acme-prod' },
  { value: 'inlq7fplr1up5atx5zpxotbabm', label: 'acme-stage' },
]

const branchOptions = [
  { value: 'brnq7fplr1up5atx5zpxotbabm', label: 'checkout / main' },
  { value: 'brn4plkdhwau58atwfd92vlc8q', label: 'billing / main' },
]

const Harness = ({
  initial,
  branchesLoading,
}: {
  initial: TPermissionEntry[]
  branchesLoading?: boolean
}) => {
  const [value, setValue] = useState(initial)

  return (
    <PermissionEntries
      value={value}
      onChange={setValue}
      appOptions={appOptions}
      installOptions={installOptions}
      branchOptions={branchesLoading ? [] : branchOptions}
      branchesLoading={branchesLoading}
      orgId={ORG_ID}
    />
  )
}

export const Empty = () => <Harness initial={[]} />

export const ReadAllWriteScoped = () => (
  <Harness
    initial={readAllWriteScoped({
      orgId: ORG_ID,
      installIds: ['inl4plkdhwau58atwfd92vlc8q'],
    })}
  />
)

export const WildcardScopedToApp = () => (
  <Harness
    initial={[
      {
        resource_type: 'install',
        resource_id: '*',
        scope_type: 'app',
        scope_id: 'app98e2wpzdxwoey393edtqj45',
        permissions: ['read', 'update'],
      },
    ]}
  />
)

export const UnknownResourceID = () => (
  <Harness
    initial={[
      {
        resource_type: 'install',
        resource_id: 'inlnotinthefetchedpage00001',
        permissions: ['read'],
      },
    ]}
  />
)

export const SpecificAppBranch = () => (
  <Harness
    initial={[
      {
        resource_type: 'app_branch',
        resource_id: 'brnq7fplr1up5atx5zpxotbabm',
        permissions: ['read'],
      },
    ]}
  />
)

export const BranchesLoading = () => (
  <Harness
    branchesLoading
    initial={[
      {
        resource_type: 'app_branch',
        resource_id: 'brnq7fplr1up5atx5zpxotbabm',
        permissions: ['read'],
      },
    ]}
  />
)

export const Invalid = () => (
  <Harness
    initial={[{ resource_type: 'install', resource_id: '*', permissions: [] }]}
  />
)
