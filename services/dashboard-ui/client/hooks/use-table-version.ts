import { useContext } from 'react'
import {
  DashboardPreferencesContext,
  DEFAULT_TABLE_VERSION,
  type TTableVersion,
} from '@/providers/dashboard-preferences-provider'

export const useTableVersion = (): TTableVersion => {
  const context = useContext(DashboardPreferencesContext)
  return context?.tableVersion ?? DEFAULT_TABLE_VERSION
}
