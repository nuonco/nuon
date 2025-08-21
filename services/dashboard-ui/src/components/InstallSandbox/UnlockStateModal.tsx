'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { LockKeyOpen } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Link } from '@/components/Link'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { trackEvent } from '@/utils'
import { unlockWorkspace } from '@/components/runner-actions'
import { jobHrefPath, jobName } from '@/components/Runners/helpers'

interface IUnlockModal {
  orgId: string
  workspace: any
  lock: any
}

export const UnlockModal: FC<IUnlockModal> = ({ orgId, workspace, lock }) => {
  const { user } = useUser()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading="Unlock the terraform workspace"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              {error ? <Notice>{error}</Notice> : null}
              {lock?.runner_job ? (
                <Text>
                  This Terraform state is associated with this{' '}
                  <Link href={`/${orgId}/${jobHrefPath(lock?.runner_job)}`}>
                    {jobName(lock?.runner_job)}
                  </Link>
                </Text>
              ) : null}
              <Text className="!leading-loose" variant="reg-14">
                Are you sure you want to unlock this terraform workspace?
              </Text>
              <div className="mt-4 flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-sm"
                >
                  Cancel
                </Button>
                <Button
                  onClick={() => {
                    setIsLoading(true)
                    unlockWorkspace({
                      orgId,
                      workspaceId: workspace.id,
                    })
                      .then(() => {
                        trackEvent({
                          event: 'terraform_workspace_state_unlock',
                          user,
                          status: 'ok',
                          props: { orgId, workspaceId: workspace.id },
                        })
                        setIsLoading(false)
                        setIsOpen(false)
                      })
                      .catch((err: any) => {
                        trackEvent({
                          event: 'terraform_workspace_state_unlock',
                          user,
                          status: 'error',
                          props: { orgId, workflowId: workspace.id, err },
                        })
                        setError(
                          'Error occured, please refresh page and try again.'
                        )
                        setIsLoading(false)

                        console.error(err)
                      })
                  }}
                  className="text-base flex items-center gap-1"
                  variant="primary"
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <>
                      <SpinnerSVG />
                      Unlocking...
                    </>
                  ) : (
                    <>
                      <LockKeyOpen size="18" />
                      Force unlock
                    </>
                  )}
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
        <LockKeyOpen size="18" />
        Force unlock
      </Button>
    </>
  )
}
