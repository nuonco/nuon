'use client'

import { useState } from 'react'
import { createPortal } from 'react-dom'
import { CodeBlockIcon } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { ClickToCopyButton } from '@/components/old/ClickToCopy'
import { JsonView } from '@/components/old/Code'
import { Loading } from '@/components/old/Loading'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'

export const InstallStateModal = () => {
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
                  View install state
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <InstallState />
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-base"
                >
                  Close
                </Button>
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
        <CodeBlockIcon size="16" />
        View state
      </Button>
    </>
  )
}

const InstallState = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: state,
    error,
    isLoading,
  } = useQuery<Record<string, any>>({
    path: `/api/orgs/${org?.id}/installs/${install?.id}/state`,
  })

  return (
    <div className="flex flex-col gap-4 mb-6">
      {error ? (
        <Notice>{error?.error || 'Unable to load install state.'}</Notice>
      ) : null}
      {isLoading ? (
        <Loading loadingText="Loading install state..." variant="stack" />
      ) : (
        <div className="flex flex-col gap-4">
          <ClickToCopyButton
            className="w-fit self-end"
            textToCopy={JSON.stringify(state)}
          />
          <div className="overflow-auto max-h-[600px]">
            <JsonView data={state} />
          </div>
        </div>
      )}
    </div>
  )
}
