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
