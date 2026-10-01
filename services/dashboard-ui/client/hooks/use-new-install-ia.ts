import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'
import { isNewInstallIAEnabled } from '@/lib/install-path'

export const useNewInstallIA = () => {
  const newAppIA = useOrgFeatureFlag('new-app-ia')

  return isNewInstallIAEnabled({
    'new-app-ia': newAppIA,
  })
}
