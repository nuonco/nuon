'use client'

import { SlidersHorizontalIcon } from '@phosphor-icons/react'
import { DriftScanButton } from "@/components/DriftScanButton"
import { Dropdown } from '@/components/Dropdown'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import type { TComponentConfig } from "@/types"
import { DeleteComponentModal } from './DeleteComponentModal'
import { InstallDeployBuildModal } from './DeployBuildModal'



function hasDriftScan(componentType: TComponentConfig['type']): boolean {
  return componentType === "helm_chart" || componentType === "terraform_module" || componentType === "kubernetes_manifest"
}

interface IInstallComponentManagementDropdown {
  componentId: string
  componentName: string
  componentType?: TComponentConfig['type']
  currentBuildId?: string
}

export const InstallComponentManagementDropdown = ({
  componentName,
  componentId,
  componentType,
  currentBuildId,
}: IInstallComponentManagementDropdown) => {
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
        {hasDriftScan(componentType) ? <DriftScanButton componentId={componentId} initBuildId={currentBuildId} /> : null}
        <InstallDeployBuildModal componentId={componentId} />
        <>
          <hr className="my-2" />
          <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
            Remove
          </Text>
          <DeleteComponentModal componentId={componentId} componentName={componentName} />
        </>
      </div>
    </Dropdown>
  ) : null
}

