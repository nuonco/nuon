import type { TAppBranchRunPreviewMode } from '@/types'

export const previewModeDisplayLabel = (
  mode: TAppBranchRunPreviewMode
): string => {
  switch (mode) {
    case 'apply':
      return 'Apply'
    case 'build-only':
      return 'Build and validate'
    default:
      return 'Plan only'
  }
}
