'use client'

import React, { type FC } from 'react'
import { GearFine } from '@phosphor-icons/react/dist/ssr'
import { BreakGlassLink } from './BreakGlassLink'
import { EditModal } from './EditModal'
import { ForgetModal } from './ForgetModal'
import { ReprovisionModal } from './ReprovisionModal'
import { Dropdown } from '@/components/Dropdown'
import { useOrg } from '@/components/Orgs'
import { Text } from '@/components/Typography'
import type { TInstall } from '@/types'

interface IInstallManagementDropdown {
  hasInstallComponents?: boolean
  install: TInstall
  orgId: string
}

export const InstallManagementDropdown: FC<IInstallManagementDropdown> = ({
  install,
  orgId,
}) => {
  const { org } = useOrg()
  return (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      alignment="right"
      id="mgmt-install"
      text={
        <>
          <GearFine size="16" />
          Configure
        </>
      }
      isDownIcon
      wrapperClassName="z-20"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Settings
        </Text>

        <EditModal install={install} orgId={orgId} />
        <BreakGlassLink installId={install.id} />

        <hr className="my-2" />
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Controls
        </Text>
        <ReprovisionModal installId={install.id} orgId={orgId} />

        <>
          <hr className="my-2" />
          <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
            Remove
          </Text>

          <ForgetModal install={install} orgId={orgId} />
        </>
      </div>
    </Dropdown>
  )
}
