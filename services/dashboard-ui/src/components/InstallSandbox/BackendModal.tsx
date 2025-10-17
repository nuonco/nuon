'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'

import { createFileDownload } from '@/utils/file-download'
import { Button } from '@/components/Button'
import { Modal } from '@/components/Modal'
import { BracketsCurly, FileArrowDown } from '@phosphor-icons/react'
import { Text, Code } from '@/components/Typography'

interface IBackendModal {
  orgId: string
  workspace: any
  token: string
}

export const BackendModal: FC<IBackendModal> = ({
  orgId,
  workspace,
  token,
}) => {
  const [isOpen, setIsOpen] = useState(false)

  const backendBlock = `
terraform {
  backend "http" {
    lock_method    = "POST"
    unlock_method  = "POST"
    address = "${process.env.NEXT_PUBLIC_API_URL}/v1/terraform-backend?workspace_id=${workspace.id}&org_id=${orgId}&token=${token}"
    lock_address = "${process.env.NEXT_PUBLIC_API_URL}/v1/terraform-workspaces/${workspace.id}/lock?org_id=${orgId}&token=${token}"
    unlock_address = "${process.env.NEXT_PUBLIC_API_URL}/v1/terraform-workspaces/${workspace.id}/unlock?org_id=${orgId}&token=${token}"
  }
}
`

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading="Use the Terraform CLI"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <Text className="!leading-loose" variant="reg-14">
                To manage the Terraform state directly, download the backend
                config, add it to your Terraform project, and run the following
                command.
              </Text>
              <Code className="mt-4">terraform init -reconfigure</Code>
              <div className="mt-4 flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-sm"
                >
                  Close
                </Button>
                <Button
                  onClick={() => {
                    createFileDownload(backendBlock, 'nuon_backend.tf')
                  }}
                  className="text-base flex items-center gap-1"
                  variant="primary"
                >
                  <FileArrowDown size="18" />
                  Download
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <BracketsCurly size="16" />
        Use Terraform CLI
      </Button>
    </>
  )
}
