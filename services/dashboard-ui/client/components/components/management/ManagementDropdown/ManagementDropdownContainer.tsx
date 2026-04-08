import { useApp } from '@/hooks/use-app'
import { ManagementDropdown } from './ManagementDropdown'

export const ManagementDropdownContainer = () => {
  const { app } = useApp()
  return (
    <ManagementDropdown
      appId={app?.id}
      appConfigId={app?.app_configs?.[0]?.id}
    />
  )
}
