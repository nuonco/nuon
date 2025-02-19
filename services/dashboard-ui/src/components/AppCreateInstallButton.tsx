'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useState } from 'react'
import { Cube } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { InstallForm } from '@/components/InstallForm'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import { createAppInstall } from './app-actions'
import type { TAppInputConfig } from '@/types'

interface IAppCreateInstallButton {
  appId: string
  platform: string | 'aws' | 'azure'
  inputConfig: TAppInputConfig
  orgId: string
}

export const AppCreateInstallButton: FC<IAppCreateInstallButton> = ({
  appId,
  inputConfig,
  orgId,
  platform,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const router = useRouter()

  return (
    <>
      <Modal
        className="!max-w-5xl"
        isOpen={isOpen}
        heading={
          <span className="flex flex-col gap-2">
            <Text variant="med-18">Create install</Text>
            <Text variant="reg-14" className="!font-normal">
              Enter the following information to setup your install.
            </Text>
          </span>
        }
        onClose={() => {
          setIsOpen(false)
        }}
        contentClassName="px-0 py-0"
      >
        <InstallForm
          onSubmit={(formData) => {
            return createAppInstall({
              appId,
              orgId,
              formData,
              platform,
            })
          }}
          onSuccess={(install) => {
            router.push(`/${orgId}/installs/${install.id}/history`)
          }}
          onCancel={() => {
            setIsOpen(false)
          }}
          platform={platform}
          inputConfig={inputConfig}
        />
      </Modal>
      <Button
        className="flex items-center gap-2 text-sm font-medium"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <Cube size={16} /> Create install
      </Button>
    </>
  )
}
