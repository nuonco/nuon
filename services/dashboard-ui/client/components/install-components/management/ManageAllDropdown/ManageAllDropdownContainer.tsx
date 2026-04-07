import { useInstall } from '@/hooks/use-install'
import { ManageAllDropdown } from './ManageAllDropdown'

export const ManageAllDropdownContainer = () => {
  const { install } = useInstall()
  return (
    <ManageAllDropdown
      appId={install?.app_id}
      appConfigId={install?.app_config_id}
    />
  )
}
