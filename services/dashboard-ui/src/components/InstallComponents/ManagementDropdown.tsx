'use client'

import React, { type FC } from 'react'
import { SlidersHorizontal } from '@phosphor-icons/react'
import { Dropdown } from '@/components/Dropdown'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import type { TComponent } from '@/types'
import { DeleteComponentModal } from './DeleteComponentModal'
import { InstallDeployBuildModal } from './DeployBuildModal'

interface IInstallComponentManagementDropdown {
  component: TComponent
}

export const InstallComponentManagementDropdown: FC<
  IInstallComponentManagementDropdown
> = ({ component }) => {
  const { org } = useOrg()
  return org?.features?.['install-delete-components'] ? (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      alignment="right"
      id="mgmt-install"
      text={
        <>
          <SlidersHorizontal />
          Manage
        </>
      }
      isDownIcon
      wrapperClassName="z-20"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Controls
        </Text>
        <InstallDeployBuildModal componentId={component?.id} />
        <>
          <hr className="my-2" />
          <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
            Remove
          </Text>
          <DeleteComponentModal component={component} />
        </>
      </div>
    </Dropdown>
  ) : null
}
