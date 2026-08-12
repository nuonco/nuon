import type {
  TPermissionEntry,
  TPermissionResourceType,
  TPermissionVerb,
} from '@/types'

export const RESOURCE_TYPE_OPTIONS: {
  value: TPermissionResourceType
  label: string
}[] = [
  { value: 'install', label: 'Installs' },
  { value: 'app', label: 'Apps' },
  { value: 'app_branch', label: 'App branches' },
  { value: 'org', label: 'This org' },
]

export const VERB_OPTIONS: { value: TPermissionVerb; label: string }[] = [
  { value: 'create', label: 'Create' },
  { value: 'read', label: 'Read' },
  { value: 'update', label: 'Update' },
  { value: 'delete', label: 'Delete' },
]

export const ALL_VERBS: TPermissionVerb[] = [
  'create',
  'read',
  'update',
  'delete',
]

const RESOURCE_TYPE_LABELS: Record<TPermissionResourceType, string> = {
  install: 'installs',
  app: 'apps',
  app_branch: 'app branches',
  org: 'org',
}

// Mirrors Level.ValidWildcardScope in services/ctl-api/internal/app/policy.go:
// only installs and branches may confine a wildcard, and only to an app.
export function canScopeWildcard(
  resourceType: TPermissionResourceType
): boolean {
  return resourceType === 'install' || resourceType === 'app_branch'
}

export function isWildcard(entry: TPermissionEntry): boolean {
  return entry.resource_id === '*'
}

export function newEntry(
  resourceType: TPermissionResourceType = 'install'
): TPermissionEntry {
  return {
    resource_type: resourceType,
    resource_id: '*',
    permissions: [...ALL_VERBS],
  }
}

// entryError returns why the API would reject an entry, or null when it is
// valid. Mirrors validatePermissionEntry so the user sees the problem inline
// rather than as a 400 after submitting.
export function entryError(entry: TPermissionEntry): string | null {
  if (entry.permissions.length === 0) {
    return 'Select at least one permission'
  }
  return null
}

export function entriesValid(entries: TPermissionEntry[]): boolean {
  // Every target select defaults to the wildcard, so an empty resource_id is
  // unreachable from the picker — it still blocks submit rather than earning a
  // message, since the API would reject it.
  return (
    entries.length > 0 &&
    entries.every(
      (e) =>
        entryError(e) === null &&
        (e.resource_type === 'org' || !!e.resource_id)
    )
  )
}

export function entrySummary(
  entry: TPermissionEntry,
  nameFor?: (id: string) => string | undefined
): string {
  const verbs =
    entry.permissions.length === ALL_VERBS.length
      ? 'Full access'
      : entry.permissions
          .map((v) => v.charAt(0).toUpperCase() + v.slice(1))
          .join(', ')

  const label = (id: string) => nameFor?.(id) ?? id

  if (entry.resource_type === 'org') {
    return `${verbs} on org-level resources`
  }
  if (!isWildcard(entry)) {
    return `${verbs} on ${label(entry.resource_id)}`
  }
  if (entry.scope_id) {
    return `${verbs} on all ${RESOURCE_TYPE_LABELS[entry.resource_type]} in ${label(entry.scope_id)}`
  }
  return `${verbs} on all ${RESOURCE_TYPE_LABELS[entry.resource_type]}`
}

// readAllWriteScoped builds the entry set behind the "read everything, write to
// specific installs" preset. The org read entry is what makes collection
// endpoints work: they authorize at the org tier, so without it the role can
// open an install it owns but cannot list installs at all.
export function readAllWriteScoped({
  orgId,
  installIds,
}: {
  orgId: string
  installIds: string[]
}): TPermissionEntry[] {
  return [
    { resource_type: 'org', resource_id: orgId, permissions: ['read'] },
    { resource_type: 'app', resource_id: '*', permissions: ['read'] },
    { resource_type: 'install', resource_id: '*', permissions: ['read'] },
    ...installIds.map((id) => ({
      resource_type: 'install' as const,
      resource_id: id,
      permissions: [...ALL_VERBS],
    })),
  ]
}

export function entriesFromRole(
  policies: { scoped_permissions?: TPermissionEntry[] }[] | undefined
): TPermissionEntry[] {
  return (policies ?? []).flatMap((p) => p.scoped_permissions ?? [])
}
