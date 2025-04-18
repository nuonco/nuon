'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { PencilSimpleLine } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { InstallForm } from '@/components/InstallForm'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { updateInstall } from '@/components/install-actions'
import type { TInstall, TAppInputConfig } from '@/types'

interface IEditModal {
  install: TInstall
  orgId: string
}

export const EditModal: FC<IEditModal> = ({ install, orgId }) => {
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [inputConfig, setInputConfig] = useState<TAppInputConfig | undefined>()
  const [error, setError] = useState<string>()
  const router = useRouter()

  useEffect(() => {
    if (isOpen) {
      fetch(
        `/api/${orgId}/apps/${install?.app_id}/input-configs/${install?.install_inputs?.at(0)?.app_input_config_id}`
      )
        .then((res) =>
          res.json().then((inputs) => {
            setInputConfig(inputs as TAppInputConfig)
            setIsLoading(false)
          })
        )
        .catch((err) => {
          setIsLoading(false)
          setError(err?.message || 'Unable to fetch app input configs')
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
              heading={`Edit install settings?`}
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
                    return updateInstall({
                      installId: install.id,
                      orgId,
                      formData,
                    })
                  }}
                  onSuccess={(workflowId) => {
                    router.push(
                      `/${orgId}/installs/${install.id}/history/${workflowId}`
                    )
                    setIsOpen(false)
                  }}
                  onCancel={() => {
                    setIsOpen(false)
                  }}
                  inputConfig={inputConfig}
                  install={install}
                />
              )}
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <PencilSimpleLine size="16" />
        Edit inputs
      </Button>
    </>
  )
}
