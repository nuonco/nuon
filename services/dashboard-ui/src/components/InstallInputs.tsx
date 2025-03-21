'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { ArrowsOutSimple } from '@phosphor-icons/react/dist/ssr'
import { Button } from '@/components/Button'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import type { TInstallInputs } from '@/types'

export interface IInstallInputs {
  currentInputs?: TInstallInputs
}

export const InstallInputs: FC<IInstallInputs> = ({ currentInputs }) => {
  return (
    <div className="divide-y">
      <div className="grid grid-cols-3 gap-4 pb-3">
        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
          Name
        </Text>
        <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
          Value
        </Text>
      </div>

      <div>
        {currentInputs ? (
          <div className="divide-y" key={currentInputs.id}>
            {currentInputs?.redacted_values
              ? Object.keys(currentInputs.redacted_values).map((key, i) => (
                  <div
                    key={`${key}-${i}`}
                    className="grid grid-cols-3 gap-4 py-3"
                  >
                    <Text className="font-mono text-sm break-all !inline truncate max-w-[200px]">
                      {key}
                    </Text>
                    <Text className="col-span-2 break-all text-sm !inline truncate max-w-[200px]">
                      {currentInputs.redacted_values[key]}
                    </Text>
                  </div>
                ))
              : currentInputs?.values &&
                Object.keys(currentInputs.values).map((key, i) => (
                  <div
                    key={`${key}-${i}`}
                    className="grid grid-cols-3 gap-4 py-3"
                  >
                    <Text className="font-mono text-sm !inline truncate max-w-[200px]">
                      {key}
                    </Text>
                    <Text className="col-span-2 break-all text-sm !inline truncate max-w-[200px]">
                      {currentInputs.values[key]}
                    </Text>
                  </div>
                ))}
          </div>
        ) : null}
      </div>
    </div>
  )
}

export const InstallInputsModal: FC<IInstallInputs> = ({ currentInputs }) => {
  const [isOpen, setIsOpen] = useState(false)
  return (
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
              <InstallInputs currentInputs={currentInputs} />
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
}
