'use client'

import { useState } from 'react'
import { createPortal } from 'react-dom'
import { DownloadSimpleIcon, FileCodeIcon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { ClickToCopyButton } from '@/components/ClickToCopy'
import { CodeBlock } from '@/components/CodeBlock'
import { SpinnerSVG, Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TFileResponse } from '@/types'
import { downloadFileOnClick } from '@/utils/file-download'

export const GenerateInstallConfigModal = () => {
  const [isOpen, setIsOpen] = useState(false)

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
              <DownloadInstallCLIConfig
                handleClose={() => {
                  setIsOpen(false)
                }}
              />
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

const DownloadInstallCLIConfig = ({ handleClose }) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const {
    data: config,
    error,
    isLoading,
  } = useQuery<TFileResponse>({
    path: `/api/orgs/${org.id}/installs/${install.id}/generate-cli-config`,
  })

  return (
    <>
      <div className="flex flex-col gap-4 mb-6">
        {error ? (
          <Notice>
            {error?.error || 'Unable to load install config TOML'}
          </Notice>
        ) : null}
        {isLoading ? (
          <Loading
            loadingText="Generating install config file..."
            variant="stack"
          />
        ) : (
          <div className="flex flex-col gap-4">
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={config?.content}
            />
            <div className="overflow-auto max-h-[600px]">
              <CodeBlock language="json">{config?.content}</CodeBlock>
            </div>
          </div>
        )}
      </div>
      <div className="flex gap-3 justify-end">
        <Button onClick={handleClose} className="text-base">
          Close
        </Button>
        {isLoading || !config?.content ? (
          <Button
            disabled={isLoading}
            className="text-sm flex items-center gap-1"
            variant="primary"
            onClick={handleClose}
          >
            <SpinnerSVG /> Download TOML
          </Button>
        ) : (
          <Button
            className="text-sm flex items-center gap-1"
            variant="primary"
            onClick={() => {
              downloadFileOnClick({
                ...config,
                callback: handleClose,
              })
            }}
          >
            <DownloadSimpleIcon size="18" /> Download TOML
          </Button>
        )}
      </div>
    </>
  )
}
