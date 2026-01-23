import type { TVCSConnectionStatus, TTheme } from '@/types'

export const getStatusTheme = (
  statusValue: TVCSConnectionStatus['status']
): TTheme => {
  switch (statusValue) {
    case 'active':
      return 'success'
    case 'suspended':
      return 'error'
    case 'unknown':
      return 'warn'
    default:
      return 'neutral'
  }
}
