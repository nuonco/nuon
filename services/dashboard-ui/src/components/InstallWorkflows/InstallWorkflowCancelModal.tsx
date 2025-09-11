'use client'

import classNames from 'classnames'
import { useParams, usePathname, useRouter } from 'next/navigation'
import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { Check, XSquare } from '@phosphor-icons/react'
import { Button, type TButtonVariant } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { cancelInstallWorkflow } from '@/components/workflow-actions'
import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflow } from '@/types'
import { trackEvent, removeSnakeCase } from '@/utils'

interface IInstallWorkflowCancelModal {
  buttonClassName?: string
  buttonVariant?: TButtonVariant
  installWorkflow: TInstallWorkflow
}

export const InstallWorkflowCancelModal: FC<IInstallWorkflowCancelModal> = ({
  buttonClassName,
  buttonVariant,
  installWorkflow,
}) => {
  const { user } = useUser()
  const { org } = useOrg()
  const pathName = usePathname()
  const params =
    useParams<Record<'org-id' | 'install-id' | 'workflow-id', string>>()
  const router = useRouter()
  const orgId = params?.['org-id']
  const installWorkflowId = installWorkflow?.id
  const [isOpen, setIsOpen] = useState<boolean>(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [hasBeenCanceled, setHasBeenCanceled] = useState(false)
  const [error, setError] = useState<string>()

  const workflowType = removeSnakeCase(installWorkflow?.type)
  const workflowPath = `/${orgId}/installs/${installWorkflow?.owner_id}/workflows/${installWorkflow?.id}`
  const historyPath = `/${orgId}/installs/${installWorkflow?.owner_id}/workflows`

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-lg"
              isOpen={isOpen}
              heading={`Cancel ${workflowType}?`}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text>
                  Are you sure you want to cancel this {workflowType} workflow?
                </Text>
              </div>
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-base"
                >
                  Cancel
                </Button>
                <Button
                  disabled={Boolean(error)}
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    cancelInstallWorkflow({ orgId, installWorkflowId }).then(
                      (res) => {
                        if (res?.error) {
                          trackEvent({
                            event: 'install_workflow_cancel',
                            status: 'error',
                            user,
                            props: {
                              workflowType: installWorkflow?.type,
                              orgId: org.id,
                              installWorkflowId,
                            },
                          })
                          console.error(res?.error)
                          setIsLoading(false)
                          setError(
                            res?.error?.error ||
                              'Error occured, please refresh page and try again.'
                          )
                        } else {
                          trackEvent({
                            event: 'install_workflow_cancel',
                            status: 'ok',
                            user,
                            props: {
                              workflowType: installWorkflow?.type,
                              orgId: org.id,
                              installWorkflowId,
                            },
                          })
                          setIsLoading(false)
                          setIsKickedOff(true)
                          if (pathName !== workflowPath && pathName !== historyPath) {
                            router.push(workflowPath)
                          }

                          setIsOpen(false)
                          setHasBeenCanceled(true)
                        }
                      }
                    )
                  }}
                  variant="danger"
                >
                  {isKickedOff ? (
                    <Check size="18" />
                  ) : isLoading ? (
                    <SpinnerSVG />
                  ) : (
                    <XSquare size="18" />
                  )}{' '}
                  Cancel {workflowType}
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        disabled={hasBeenCanceled}
        className={classNames('text-sm !font-medium w-fit', {
          'text-red-800 dark:text-red-500': !hasBeenCanceled,
          'text-red-800/50 dark:text-red-500/50': hasBeenCanceled,
          [`${buttonClassName}`]: Boolean(buttonClassName),
        })}
        onClick={() => {
          setIsOpen(true)
        }}
        variant={buttonVariant}
      >
        Cancel {workflowType}
      </Button>
    </>
  )
}
