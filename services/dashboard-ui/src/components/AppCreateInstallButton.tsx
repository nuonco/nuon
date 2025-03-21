'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Cube } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { composeCloudFormationQuickCreateUrl } from '@/components/Installs/helpers'
import { InstallForm } from '@/components/InstallForm'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { createAppInstall } from './app-actions'
import type { TAppInputConfig, TAppSandboxConfig } from '@/types'

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
  const [isLoading, setIsLoading] = useState(true)
  const [sandboxConfig, setSandboxConfig] = useState<
    TAppSandboxConfig | undefined
  >()
  const [error, setError] = useState<string>()
  const router = useRouter()

  useEffect(() => {
    if (isOpen) {
      fetch(`/api/${orgId}/apps/${appId}/sandbox-configs/latest`)
        .then((res) =>
          res.json().then((sandbox) => {
            setSandboxConfig(sandbox as TAppSandboxConfig)
            setIsLoading(false)
          })
        )
        .catch((err) => {
          setIsLoading(false)
          setError(err?.message || 'Unable to fetch latest sandbox config')
        })
    }
  }, [isOpen])

  return (
    <>
      {isOpen
        ? createPortal(
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
              {isLoading ? (
                <div className="p-6">
                  <Loading loadingText="Loading configs..." variant="stack" />
                </div>
              ) : error ? (
                <div className="p-6">
                  <Notice>{error}</Notice>
                </div>
              ) : (
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
                  cfLink={composeCloudFormationQuickCreateUrl(sandboxConfig)}
                />
              )}
            </Modal>,
            document.body
          )
        : null}
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
