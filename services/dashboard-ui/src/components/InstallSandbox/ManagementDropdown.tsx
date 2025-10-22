'use client'

import { SlidersHorizontalIcon } from '@phosphor-icons/react'
import { DriftScanSandboxModal } from '@/components/InstallSandbox/SandboxDriftScanButton'
import { Dropdown } from '@/components/Dropdown'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { DeprovisionSandboxModal } from '@/components/Installs/DeprovisionSandboxModal'
import { ReprovisionSandboxModal } from '@/components/Installs/ReprovisionSandboxModal'

interface ISandboxManagementDropdown {}

export const SandboxManagementDropdown = ({}: ISandboxManagementDropdown) => {
  const { org } = useOrg()
  return org?.features?.['install-delete-components'] ? (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      alignment="right"
      id="mgmt-install"
      text={
        <>
          <SlidersHorizontalIcon />
          Manage
        </>
      }
      isDownIcon
      wrapperClassName="z-10"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Controls
        </Text>
        <DriftScanSandboxModal />
        <ReprovisionSandboxModal />
        <>
          <hr className="my-2" />
          <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
            Remove
          </Text>
          <DeprovisionSandboxModal />
        </>
      </div>
    </Dropdown>
  ) : null
}
