'use client'

import { SlidersHorizontalIcon } from '@phosphor-icons/react/dist/ssr'
import { DeployComponentsModal } from '../InstallComponents/DeployComponentsModal'
import { DeleteComponentsModal } from '../InstallComponents/DeleteComponentsModal'
import { Dropdown } from '@/components/Dropdown'
import { useOrg } from '@/hooks/use-org'

export const InstallComponentsManagementDropdown = () => {
  const { org } = useOrg()

  return (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      alignment="right"
      id="mgmt-install"
      text={
        <>
          <SlidersHorizontalIcon size="16" />
          Manage
        </>
      }
      isDownIcon
      wrapperClassName="z-18"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <DeployComponentsModal />

        {org?.features?.['install-delete-components'] ? (
          <DeleteComponentsModal />
        ) : null}
      </div>
    </Dropdown>
  )
}
