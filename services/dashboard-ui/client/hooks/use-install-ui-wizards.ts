import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'

export const useInstallUIWizards = () => useOrgFeatureFlag('install-ui-wizards')
