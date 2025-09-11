'use client'

import { useParams } from 'next/navigation'
import React, { type FC } from 'react'
import { SlidersHorizontal } from '@phosphor-icons/react/dist/ssr'
import { DeployComponentsModal } from '../InstallComponents/DeployComponentsModal'
import { DeleteComponentsModal } from '../InstallComponents/DeleteComponentsModal'
import { Dropdown } from '@/components/Dropdown'
import { useOrg } from '@/hooks/use-org'

interface IInstallComponentsManagementDropdown {}

export const InstallComponentsManagementDropdown: FC<
  IInstallComponentsManagementDropdown
> = ({}) => {
  const params =
    useParams<Record<'org-id' | 'install-id' | 'component-id', string>>()
  const { org } = useOrg()
  const installId = params['install-id']

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
      wrapperClassName="z-18"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <DeployComponentsModal installId={installId} orgId={org?.id} />

        {org?.features?.['install-delete-components'] ? (
          <DeleteComponentsModal installId={installId} orgId={org?.id} />
        ) : null}
      </div>
    </Dropdown>
  )
}
