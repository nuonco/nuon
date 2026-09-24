import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'

export const useNewInstallIA = () => {
  const appBranchesUI = useOrgFeatureFlag('app-branches-ui')
  const newInstallIA = useOrgFeatureFlag('new-install-ia')

  return appBranchesUI && newInstallIA
}
