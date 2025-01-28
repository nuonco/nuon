'use client'

import React, { type FC, useState } from 'react'
import { ArrowURightUp, CloudArrowUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { InstallDeployComponentButton } from '@/components/InstallDeployComponentsButton'
import { InstallReprovisionButton } from '@/components/InstallReprovisionButton'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'

interface IInstallManagementDropdown {
  hasInstallComponents?: boolean
  installId: string
  orgId: string
}

export const InstallManagementDropdown: FC<IInstallManagementDropdown> = ({
  hasInstallComponents = false,
  installId,
  orgId,
}) => {
  const [isDeploymentOpen, setIsDeploymentOpen] = useState(false)
  const [isReprovisionOpen, setIsReprovisionOpen] = useState(false)

  return (
    <>
      <Modal
        className="max-w-lg"
        heading="Deploy all components?"
        isOpen={isDeploymentOpen}
        onClose={() => {
          setIsDeploymentOpen(false)
        }}
      >
        <div className="mb-6">
          <Text variant="reg-14" className="leading-relaxed">
            Are you sure you want to deploy components? This will deploy all
            components to this install.
          </Text>
        </div>
        <div className="flex gap-3 justify-end">
          <Button
            onClick={() => {
              setIsDeploymentOpen(false)
            }}
            className="text-base"
          >
            Cancel
          </Button>
          <InstallDeployComponentButton
            installId={installId}
            orgId={orgId}
            onComplete={() => {
              setIsDeploymentOpen(false)
            }}
          />
        </div>
      </Modal>
      <Modal
        className="max-w-lg"
        heading="Reprovision install?"
        isOpen={isReprovisionOpen}
        onClose={() => {
          setIsReprovisionOpen(false)
        }}
      >
        <div className="mb-6">
          <Text variant="reg-14" className="leading-relaxed">
            Are you sure you want to reprovision this install?
          </Text>
        </div>
        <div className="flex gap-3 justify-end">
          <Button
            onClick={() => {
              setIsReprovisionOpen(false)
            }}
            className="text-base"
          >
            Cancel
          </Button>
          <InstallReprovisionButton
            installId={installId}
            orgId={orgId}
            onComplete={() => {
              setIsReprovisionOpen(false)
            }}
          />
        </div>
      </Modal>
      <Dropdown
        className="text-sm !font-medium !p-2 h-[32px]"
        alignment="right"
        id="mgmt-install"
        text="Admin"
        isDownIcon
        wrapperClassName="z-20"
      >
        <div className="min-w-[180px] rounded-md overflow-hidden">
          <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
            Controls
          </Text>
          <Button
            className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3 !rounded-none w-full"
            variant="ghost"
            onClick={() => {
              setIsReprovisionOpen(true)
            }}
          >
            <ArrowURightUp size="18" />
            Reprovision install
          </Button>
          {hasInstallComponents ? (
            <Button
              className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3 !rounded-none w-full"
              variant="ghost"
              onClick={() => {
                setIsDeploymentOpen(true)
              }}
            >
              <CloudArrowUp size="18" />
              Deploy components
            </Button>
          ) : null}
        </div>
      </Dropdown>
    </>
  )
}

//
