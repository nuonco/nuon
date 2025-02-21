'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useState } from 'react'
import {
  Axe,
  ArrowURightUp,
  CloudArrowUp,
  PencilSimpleLine,
  Trash,
  WarningOctagon,
} from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { ForgetInstallButton } from '@/components/ForgetInstallButton'
import { InstallDeployComponentButton } from '@/components/InstallDeployComponentsButton'
import { InstallForm } from '@/components/InstallForm'
import { InstallReprovisionButton } from '@/components/InstallReprovisionButton'
import { TeardownAllComponentsButton } from '@/components/TeardownAllComponentsButton'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import { updateInstall } from '@/components/install-actions'
import type { TInstall, TAppInputConfig } from '@/types'

interface IInstallManagementDropdown {
  hasInstallComponents?: boolean
  installId: string
  orgId: string
  install: TInstall
  inputConfig?: TAppInputConfig
  hasUpdateInstall?: boolean
}

export const InstallManagementDropdown: FC<IInstallManagementDropdown> = ({
  hasInstallComponents = false,
  installId,
  install,
  inputConfig,
  orgId,
  hasUpdateInstall = false,
}) => {
  const [isDeploymentOpen, setIsDeploymentOpen] = useState(false)
  const [isReprovisionOpen, setIsReprovisionOpen] = useState(false)
  const [isTeardownOpen, setIsTeardownOpen] = useState(false)
  const [isEditOpen, setIsEditOpen] = useState(false)
  const [isForgetOpen, setIsForgetOpen] = useState(false)
  const router = useRouter()

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
      <Modal
        className="max-w-lg"
        isOpen={isTeardownOpen}
        heading={`Teardown all components?`}
        onClose={() => {
          setIsTeardownOpen(false)
        }}
      >
        <div className="mb-6">
          <Text variant="reg-14" className="leading-relaxed">
            Are you sure you want to teardown all components? This will remove
            all components from this install but leave the sandbox and runner.
          </Text>
        </div>
        <div className="flex gap-3 justify-end">
          <Button
            onClick={() => {
              setIsTeardownOpen(false)
            }}
            className="text-base"
          >
            Cancel
          </Button>
          <TeardownAllComponentsButton
            installId={installId}
            orgId={orgId}
            onComplete={() => {
              setIsTeardownOpen(false)
            }}
          />
        </div>
      </Modal>

      <Modal
        className="!max-w-5xl"
        isOpen={isEditOpen}
        heading={`Edit install settings?`}
        onClose={() => {
          setIsEditOpen(false)
        }}
        contentClassName="px-0 py-0"
      >
        <InstallForm
          onSubmit={(formData) => {
            return updateInstall({
              installId,
              orgId,
              formData,
            })
          }}
          onSuccess={(install) => {
            router.push(`/${orgId}/installs/${install.id}/history`)
          }}
          onCancel={() => {
            setIsEditOpen(false)
          }}
          inputConfig={inputConfig}
          install={install}
        />
      </Modal>

      <Modal
        className="max-w-lg"
        isOpen={isForgetOpen}
        heading={
          <span className="flex items-center gap-3">
            Forget {install.name}?
          </span>
        }
        onClose={() => {
          setIsForgetOpen(false)
        }}
      >
        <div className="flex flex-col gap-4 mb-6">
          <span className="flex items-center gap-3 w-full py-2.5 pr-4 pl-2 border rounded-md border-red-400 bg-red-300/20 text-red-800 dark:border-red-600 dark:bg-red-600/5 dark:text-red-600 text-sm font-medium leading-normal">
            <WarningOctagon size={30} /> This should only be used in cases where
            an install was broken in an unordinary way and needs to be manually
            removed.
          </span>
          <Text variant="reg-14" className="leading-relaxed">
            Are you sure you want to forget {install?.name}? <br /> This action
            will remove the install and can not be undone.
          </Text>
        </div>
        <div className="flex gap-3 justify-end">
          <Button
            onClick={() => {
              setIsForgetOpen(false)
            }}
            className="text-base"
          >
            Cancel
          </Button>
          <ForgetInstallButton
            installId={installId}
            orgId={orgId}
            onComplete={() => {
              router.push(`/${orgId}/installs`)
              setIsForgetOpen(false)
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
          {hasUpdateInstall ? (
            <Button
              className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3 !rounded-none w-full"
              variant="ghost"
              onClick={() => {
                setIsEditOpen(true)
              }}
            >
              <PencilSimpleLine size="18" />
              Edit install
            </Button>
          ) : null}
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
          {hasInstallComponents && hasUpdateInstall ? (
            <Button
              className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3 !rounded-none w-full"
              variant="ghost"
              onClick={() => {
                setIsTeardownOpen(true)
              }}
            >
              <Axe size="18" /> Teardown components
            </Button>
          ) : null}
          {hasUpdateInstall ? (
            <>
              <hr />
              <Button
                className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3 !rounded-none w-full text-red-800 dark:text-red-500"
                variant="ghost"
                onClick={() => {
                  setIsForgetOpen(true)
                }}
              >
                <Trash size="18" />
                Forget install
              </Button>
            </>
          ) : null}
        </div>
      </Dropdown>
    </>
  )
}
