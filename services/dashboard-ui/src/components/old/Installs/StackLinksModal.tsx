'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { Link as LinkIcon } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { ClickToCopyButton } from '@/components/old/ClickToCopy'
import { Link } from '@/components/old/Link'
import { Modal } from '@/components/old/Modal'
import { Code, Text } from '@/components/old/Typography'

interface IStackLinksModal {
  template_url: string
  quick_link_url: string
}

export const StackLinksModal: FC<IStackLinksModal> = ({
  template_url,
  quick_link_url,
}) => {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-3xl"
              contentClassName="!max-h-max"
              isOpen={isOpen}
              heading={
                <span className="flex items-center gap-3">
                  View stack links
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-6">
                <div className="border rounded-md shadow p-2 flex flex-col gap-1">
                  <span className="flex justify-between items-center">
                    <Text variant="med-12">Install quick link</Text>
                    <ClickToCopyButton textToCopy={quick_link_url} />
                  </span>
                  <Link
                    href={quick_link_url}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Code>{quick_link_url}</Code>
                  </Link>
                </div>

                <div className="border rounded-md shadow p-2 flex flex-col gap-1 mt-3">
                  <span className="flex justify-between items-center">
                    <Text variant="med-12">Install template link</Text>
                    <ClickToCopyButton textToCopy={template_url} />
                  </span>
                  <Link
                    href={template_url}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Code>{template_url}</Code>
                  </Link>
                </div>
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
        <LinkIcon size="16" />
        View links
      </Button>
    </>
  )
}
