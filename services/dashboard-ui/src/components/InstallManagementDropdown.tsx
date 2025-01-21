import React, { type FC } from 'react'
import { Dropdown } from '@/components/Dropdown'
import { InstallDeployComponentButton } from '@/components/InstallDeployComponentsButton'
import { InstallReprovisionButton } from '@/components/InstallReprovisionButton'
import { Text } from '@/components/Typography'

interface IInstallManagementDropdown {
  hasInstallComponents?: boolean
  installId: string
  orgId: string
}

export const InstallManagementDropdown: FC<IInstallManagementDropdown> = ({
  hasInstallComponents = false,
  installId,
  orgId,
}) => {
  return (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      alignment="right"
      id="mgmt-install"
      text="Admin"
      isDownIcon
    >
      <div className="min-w-[180px] rounded-md overflow-hidden">
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Controls
        </Text>
        <InstallReprovisionButton installId={installId} orgId={orgId} />
        {hasInstallComponents ? (
          <InstallDeployComponentButton installId={installId} orgId={orgId} />
        ) : null}
      </div>
    </Dropdown>
  )
}
