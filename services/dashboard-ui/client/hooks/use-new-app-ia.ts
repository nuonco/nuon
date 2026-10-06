import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'

export const useNewAppIA = () => useOrgFeatureFlag('new-app-ia')
