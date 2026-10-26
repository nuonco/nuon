import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'

export const usePRPreviewUI = () => useOrgFeatureFlag('pr-preview-ui')
