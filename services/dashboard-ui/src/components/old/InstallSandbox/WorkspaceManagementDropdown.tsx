'use client'

import React, { type FC } from 'react'
import { SlidersHorizontalIcon } from '@phosphor-icons/react'
import { Dropdown } from '@/components/old/Dropdown'
import { BackendModal } from '@/components/old/InstallSandbox/BackendModal'

interface IWorkspaceManagementDropdown {
  workspace: any
  orgId: string
  token: string
}

export const WorkspaceManagementDropdown: FC<IWorkspaceManagementDropdown> = ({
  workspace,
  orgId,
  token,
}) => {
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
      wrapperClassName="z-20"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <BackendModal orgId={orgId} workspace={workspace} token={token} />
      </div>
    </Dropdown>
  )
}
