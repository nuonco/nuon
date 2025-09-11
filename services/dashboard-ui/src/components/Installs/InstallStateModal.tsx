'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { CodeBlock } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { ClickToCopyButton } from '@/components/ClickToCopy'
import { JsonView } from '@/components/Code'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { useOrg } from '@/hooks/use-org'
import type { TInstall } from '@/types'

interface IInstallStateModal {
  install: TInstall
}

export const InstallStateModal: FC<IInstallStateModal> = ({ install }) => {
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [state, setState] = useState()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState()

  useEffect(() => {
    if (isOpen) {
      fetch(`/api/${org?.id}/installs/${install?.id}/state`).then((r) =>
        r.json().then((res) => {
          setIsLoading(false)
          if (res?.error) {
            setError(res?.error?.error || 'Unable to fetch install state')
          } else {
            setState(res.data)
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
              <div className="flex flex-col gap-4 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                {isLoading ? (
                  <Loading
                    loadingText="Loading install state..."
                    variant="stack"
                  />
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
        <CodeBlock size="16" />
        View state
      </Button>
    </>
  )
}
