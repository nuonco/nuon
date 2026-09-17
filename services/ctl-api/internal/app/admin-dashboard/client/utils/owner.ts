// Polymorphic owner_type / owner_id pairs show up on queues, signals,
// workflows and log streams. Both maps resolve an owner to the admin page that
// renders it; the ID prefix map covers rows whose owner_type is empty.

const OWNER_TYPE_PATHS: Record<string, string> = {
  installs: 'installs',
  app_branches: 'app-branches',
  orgs: 'orgs',
  accounts: 'accounts',
  runners: 'runners',
  install_workflows: 'workflows',
}

const ID_PREFIX_PATHS: Record<string, string> = {
  inl: 'installs',
  abr: 'app-branches',
  org: 'orgs',
  acc: 'accounts',
  run: 'runners',
  inw: 'workflows',
}

export function ownerPath(ownerType?: string, ownerId?: string): string | null {
  if (!ownerId) return null
  const segment = (ownerType && OWNER_TYPE_PATHS[ownerType]) || ID_PREFIX_PATHS[ownerId.slice(0, 3)]
  return segment ? `/${segment}/${ownerId}` : null
}
