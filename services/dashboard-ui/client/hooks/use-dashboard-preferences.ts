import { useContext } from 'react'
import { DashboardPreferencesContext } from '@/providers/dashboard-preferences-provider'

export const useDashboardPreferences = () => {
  const context = useContext(DashboardPreferencesContext)
  if (!context) {
    throw new Error(
      'useDashboardPreferences must be used within a DashboardPreferencesProvider'
    )
  }
  return context
}
