'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { DownloadSimpleIcon, FileCodeIcon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { ClickToCopyButton } from '@/components/ClickToCopy'
import { CodeViewer } from '@/components/Code'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'

interface IGenerateInstallConfigButton {
  installId: string
  orgId: string
}

export const GenerateInstallConfigModal: FC<IGenerateInstallConfigButton> = ({
  installId,
  orgId,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const [config, setConfig] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState()

  useEffect(() => {
    if (isOpen) {
      try {
        fetch(
          `/api/${orgId}/installs/${installId}/generate-cli-install-config`
        ).then((r) =>
          r.text().then((res) => {
            setIsLoading(false)
            setConfig(res)
          })
        )
      } catch (err) {
        setError(err)
      }
    }
  }, [isOpen])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-5xl"
              contentClassName="!max-h-max"
              isOpen={isOpen}
              heading={
                <span className="flex items-center gap-3">
                  Generate Install Config
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                {isLoading ? (
                  <Loading
                    loadingText="Generating install config file..."
                    variant="stack"
                  />
                ) : (
                  <div className="flex flex-col gap-4">
                    <ClickToCopyButton
                      className="w-fit self-end"
                      textToCopy={config}
                    />
                    <div className="overflow-auto max-h-[600px]">
                      <CodeViewer initCodeSource={config} language="toml" />
                    </div>
                  </div>
                )}
              </div>
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-base"
                >
                  Close
                </Button>
                <a
                  href={`/api/${orgId}/installs/${installId}/generate-cli-install-config`}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <Button
                    className="flex items-center gap-1 text-sm font-medium disabled:!bg-primary-950"
                    type="submit"
                    variant="primary"
                  >
                    <DownloadSimpleIcon size="18" /> Download File
                  </Button>
                </a>
              </div>
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
        <FileCodeIcon size="16" />
        Generate Install Config
      </Button>
    </>
  )
}
