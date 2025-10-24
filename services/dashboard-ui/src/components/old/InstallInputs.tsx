'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { ArrowsOutSimple } from '@phosphor-icons/react/dist/ssr'
import { Button } from '@/components/old/Button'
import { ConfigVariables } from '@/components/old/ComponentConfig'
import { Modal } from '@/components/old/Modal'
import type { TInstallInputs } from '@/types'

export interface IInstallInputs {
  currentInputs?: TInstallInputs
}

export const InstallInputs: FC<IInstallInputs> = ({ currentInputs }) => {
  const variables = currentInputs?.redacted_values || {}
  const variableKeys = Object.keys(variables)
  const isEmpty = variableKeys.length === 0

  return (
    !isEmpty && <ConfigVariables variables={variables} keys={variableKeys} />
  )
}

export const InstallInputsModal: FC<IInstallInputs> = ({ currentInputs }) => {
  const variables = currentInputs?.redacted_values || {}
  const variableKeys = Object.keys(variables)
  const isEmpty = variableKeys.length === 0
  const [isOpen, setIsOpen] = useState(false)

  return (
    !isEmpty && (
      <>
        {isOpen
          ? createPortal(
              <Modal
                heading="Current install inputs"
                isOpen={isOpen}
                onClose={() => {
                  setIsOpen(false)
                }}
              >
                <ConfigVariables
                  variables={variables}
                  keys={variableKeys}
                  isNotTruncated
                />
              </Modal>,
              document.body
            )
          : null}
        <Button
          className="text-sm !font-medium flex items-center gap-2 !p-1"
          onClick={() => {
            setIsOpen(true)
          }}
          title="Expand install inputs"
          variant="ghost"
        >
          <ArrowsOutSimple />
        </Button>
      </>
    )
  )
}
