import React, { type FC } from 'react'
import { BreakGlassLink } from './BreakGlassLink'
import { DeployComponentsModal } from './DeployComponentsModal'
import { EditModal } from './EditModal'
import { ForgetModal } from './ForgetModal'
import { ReprovisionModal } from './ReprovisionModal'
import { TeardownComponentsModal } from './TeardownComponentsModal'
import { Dropdown } from '@/components/Dropdown'
import { Text } from '@/components/Typography'
import type { TInstall } from '@/types'

interface IInstallManagementDropdown {
  hasInstallComponents?: boolean
  install: TInstall
  orgId: string
}

export const InstallManagementDropdown: FC<IInstallManagementDropdown> = ({
  hasInstallComponents = false,
  install,
  orgId,
}) => {
  return (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      alignment="right"
      id="mgmt-install"
      text="Admin"
      isDownIcon
      wrapperClassName="z-20"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Controls
        </Text>

        <EditModal install={install} orgId={orgId} />
        <ReprovisionModal installId={install.id} orgId={orgId} />
        {hasInstallComponents ? (
          <DeployComponentsModal installId={install.id} orgId={orgId} />
        ) : null}
        {hasInstallComponents ? (
          <TeardownComponentsModal installId={install.id} orgId={orgId} />
        ) : null}

        <>
          <hr className="my-2" />
          <BreakGlassLink installId={install.id} />
          <ForgetModal install={install} orgId={orgId} />
        </>
      </div>
    </Dropdown>
  )
}
