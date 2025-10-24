'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { ListMagnifyingGlass } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { ClickToCopyButton } from '@/components/old/ClickToCopy'
import { ConfigVariables } from '@/components/old/ComponentConfig'
import { Modal } from '@/components/old/Modal'
import { Text } from '@/components/old/Typography'
import type { TInstallStackVersionRun } from '@/types'

interface IStackOutputsModal {
  runs: Array<TInstallStackVersionRun>
}

export const StackOutputsModal: FC<IStackOutputsModal> = ({ runs }) => {
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
                  View stack outputs
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-6">
                {runs?.map((run, i) => (
                  <div key={run?.id} className="flex flex-col gap-4">
                    <Text className="flex items-center justify-between">
                      <Text variant="med-14">Run {i + 1}</Text>
                      <ClickToCopyButton
                        className="w-fit self-end"
                        textToCopy={JSON.stringify(run.data_contents)}
                      />
                    </Text>
                    <div className="overflow-auto max-h-[600px]">
                      <ConfigVariables
                        keys={Object.keys(run?.data)}
                        variables={run?.data_contents as Record<string, string>}
                        isNotTruncated
                      />
                    </div>
                  </div>
                ))}
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
        <ListMagnifyingGlass size="16" />
        View outputs
      </Button>
    </>
  )
}
