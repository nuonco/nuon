'use client'

import { SlidersHorizontalIcon } from '@phosphor-icons/react/dist/ssr'
import { AutoApproveModal } from './AutoApproveModal'
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
import { useInstall } from '@/hooks/use-install'
import { GenerateInstallConfigModal } from './GenerateInstallConfigModal'

export const InstallManagementDropdown = () => {
  const { install } = useInstall()

  return (
    <Dropdown
      className="text-sm !font-medium !p-2 !h-auto"
      alignment="right"
      id="mgmt-install"
      text={
        <>
          <SlidersHorizontalIcon size="16" />
          Manage
        </>
      }
      isDownIcon
      wrapperClassName=""
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Settings
        </Text>

        {install?.install_inputs?.length ? <EditModal /> : null}

        <InstallAuditHistoryModal />
        <InstallStateModal />
        <AutoApproveModal />
        <GenerateInstallConfigModal />

        <hr className="my-2" />
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Controls
        </Text>
        <ReprovisionModal />
        <SyncSecretsModal />
        <DeleteInstallModal />
        <DeprovisionStackModal />

        <hr className="my-2" />
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Remove
        </Text>
        <ForgetModal />
      </div>
    </Dropdown>
  )
}
