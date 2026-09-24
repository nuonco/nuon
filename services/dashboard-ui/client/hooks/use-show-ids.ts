import { useContext } from 'react'
import { DashboardPreferencesContext } from '@/providers/dashboard-preferences-provider'

export const useShowIds = (): boolean => {
  const context = useContext(DashboardPreferencesContext)
  return context?.showIds ?? false
}
