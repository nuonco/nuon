import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'
import { isNewInstallIAEnabled } from '@/lib/install-path'

export const useNewInstallIA = () => {
  const appBranchesUI = useOrgFeatureFlag('app-branches-ui')
  const newInstallIA = useOrgFeatureFlag('new-install-ia')

  return isNewInstallIAEnabled({
    'app-branches-ui': appBranchesUI,
    'new-install-ia': newInstallIA,
  })
}
