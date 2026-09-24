import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'

export const useNewAppIA = () => {
  const appBranchesUI = useOrgFeatureFlag('app-branches-ui')
  const newAppIA = useOrgFeatureFlag('new-app-ia')

  return appBranchesUI && newAppIA
}
