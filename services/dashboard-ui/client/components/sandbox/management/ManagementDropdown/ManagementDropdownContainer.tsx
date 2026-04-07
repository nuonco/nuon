import { useInstall } from '@/hooks/use-install'
import { ManagementDropdown } from './ManagementDropdown'
import type { IDropdown } from '@/components/common/Dropdown'

export const ManagementDropdownContainer = ({
  alignment = 'right',
  ...props
}: Omit<IDropdown, 'id' | 'buttonText' | 'children'>) => {
  const { install } = useInstall()
  const workspaceId = install?.sandbox?.terraform_workspace?.id

  return (
    <ManagementDropdown
      alignment={alignment}
      workspaceId={workspaceId}
      {...props}
    />
  )
}
