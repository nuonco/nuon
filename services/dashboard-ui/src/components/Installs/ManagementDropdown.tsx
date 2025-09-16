'use client'

import React, { type FC } from 'react'
import { SlidersHorizontal } from '@phosphor-icons/react/dist/ssr'
import { AutoApproveModal } from './AutoApproveModal'
import { BreakGlassLink } from './BreakGlassLink'
import { DeleteInstallModal } from './DeleteModal'
import { DeprovisionStackModal } from './DeprovisionStackModal'
import { EditModal } from './EditModal'
import { ForgetModal } from './ForgetModal'
import { InstallStateModal } from './InstallStateModal'
import { ReprovisionModal } from './ReprovisionModal'
import { InstallAuditHistoryModal } from './InstallAuditHistoryModal'
import { SyncSecretsModal } from './SyncSecretsModal'
import { Dropdown } from '@/components/Dropdown'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import type { TInstall } from '@/types'
import { GenerateInstallConfigModal } from './GenerateInstallConfigModal'

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
          <SlidersHorizontal size="16" />
          Manage
        </>
      }
      isDownIcon
      wrapperClassName="z-20"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Settings
        </Text>

        {install?.install_inputs?.length ? (
          <EditModal install={install} orgId={orgId} />
        ) : null}
        <BreakGlassLink installId={install.id} />
        <InstallAuditHistoryModal installId={install.id} orgId={org.id} />
        <InstallStateModal install={install} />
        <AutoApproveModal install={install} />
        <GenerateInstallConfigModal installId={install.id} orgId={org.id} />

        <hr className="my-2" />
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Controls
        </Text>
        <ReprovisionModal installId={install.id} orgId={orgId} />
        <SyncSecretsModal installId={install.id} orgId={orgId} />
        <DeleteInstallModal install={install} />
        <DeprovisionStackModal install={install} orgId={orgId} />

        <hr className="my-2" />
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Remove
        </Text>
        <ForgetModal install={install} orgId={orgId} />
      </div>
    </Dropdown>
  )
}
