import { useOrg } from '@/hooks/use-org'
import type { TInstall } from '@/types'
import { resolveEnvAccent } from '@/utils/env-accent'

/**
 * Convenience hook that resolves an env accent for the given install using
 * the current org's mapping. Pass the install directly so the hook works
 * for table rows where no InstallProvider is in scope.
 *
 * Returns null when no accent applies.
 */
export const useEnvAccent = (install?: Pick<TInstall, 'labels'> | null) => {
  const { org } = useOrg()
  return resolveEnvAccent(install, org)
}
