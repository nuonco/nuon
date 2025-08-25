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
import { updateInstall, updateInstallManagedBy } from '@/components/install-actions'
import type { TInstall, TAppInputConfig } from '@/types'
import { ConfirmUpdateModal } from './ConfirmUpdateModal'

interface IEditModal {
  install: TInstall
  orgId: string
}

export const EditModal: FC<IEditModal> = ({ install, orgId }) => {
  const [isOpen, setIsOpen] = useState(false)
  const [isConfirmOpen, setIsConfirmOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [inputConfig, setInputConfig] = useState<TAppInputConfig | undefined>()
  const [error, setError] = useState<string>()
  const router = useRouter()

  useEffect(() => {
    if (isOpen) {
      fetch(`/api/${orgId}/apps/${install?.app_id}/input-configs/latest`).then(
        (r) =>
          r.json().then((res) => {
            setIsLoading(false)
            if (res?.error) {
              setError(res?.error?.error || 'Unable to fetch app input configs')
            } else {
              setInputConfig(res.data)
            }
          })
      )
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
                    const res = updateInstall({
                      installId: install.id,
                      orgId,
                      formData,
                    })
                    updateInstallManagedBy({
                      installId: install?.id,
                      orgId: orgId,
                      managedBy: install?.metadata?.managed_by,
                    })
                    return res
                  }}
                  onSuccess={(workflowId) => {
                    if (workflowId) {
                      router.push(
                        `/${orgId}/installs/${install.id}/workflows/${workflowId}`
                      )
                    } else {
                      router.push(`/${orgId}/installs/${install.id}/workflows`)
                    }
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
      <ConfirmUpdateModal
        install={install}
        isOpen={isConfirmOpen}
        onClose={(isConfirmed) => {
          setIsOpen(isConfirmed)
          setIsConfirmOpen(false)
        }}
      />
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        variant="ghost"
        onClick={() => {
          setIsConfirmOpen(true)
        }}
      >
        <PencilSimpleLine size="16" />
        Edit inputs
      </Button>
    </>
  )
}
