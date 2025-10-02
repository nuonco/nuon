'use client'

import classNames from 'classnames'
import { usePathname, useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, XSquareIcon } from '@phosphor-icons/react'
import { cancelWorkflow } from '@/actions/workflows/cancel-workflow'
import { Button, type TButtonVariant } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'
import type { TWorkflow } from '@/types'
import { removeSnakeCase } from '@/utils'

interface IInstallWorkflowCancelModal {
  buttonClassName?: string
  buttonVariant?: TButtonVariant
  installWorkflow: TWorkflow
}

export const InstallWorkflowCancelModal: FC<IInstallWorkflowCancelModal> = ({
  buttonClassName,
  buttonVariant,
  installWorkflow,
}) => {
  const path = usePathname()
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()

  const installWorkflowId = installWorkflow?.id
  const [isOpen, setIsOpen] = useState<boolean>(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [hasBeenCanceled, setHasBeenCanceled] = useState(false)

  const { data, error, execute, headers, isLoading } = useServerAction({
    action: cancelWorkflow,
  })

  const workflowType = removeSnakeCase(installWorkflow?.type)
  const workflowPath = `/${org.id}/installs/${installWorkflow?.owner_id}/workflows/${installWorkflow?.id}`
  const historyPath = `/${org.id}/installs/${installWorkflow?.owner_id}/workflows`

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'workflow_cancel',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          workflowId: installWorkflowId,
          workflowType: installWorkflow?.type,
        },
      })
    }

    if (data) {
      setHasBeenCanceled(true)
      trackEvent({
        event: 'workflow_cancel',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          workflowId: installWorkflowId,
          workflowType: installWorkflow?.type,
        },
      })

      if (path !== workflowPath && path !== historyPath) {
        router.push(workflowPath)
      }

      setIsOpen(false)
    }
  }, [data, error, headers])

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
                {error ? (
                  <Notice>
                    {error?.error || 'Unable to cancel workflow.'}
                  </Notice>
                ) : null}
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
                    setIsKickedOff(true)
                    execute({ orgId: org.id, workflowId: installWorkflowId })
                  }}
                  variant="danger"
                >
                  {isLoading ? (
                    <SpinnerSVG />
                  ) : isKickedOff ? (
                    <CheckIcon size="18" />
                  ) : (
                    <XSquareIcon size="18" />
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
