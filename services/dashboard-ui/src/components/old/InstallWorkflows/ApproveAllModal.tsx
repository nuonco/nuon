'use client'

import classNames from 'classnames'
import { usePathname, useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon } from '@phosphor-icons/react'
import { approveAllWorkflowSteps } from '@/actions/workflows/approve-all-workflow-steps'
import { Badge } from '@/components/old/Badge'
import { Button, type TButtonVariant } from '@/components/old/Button'
import { SpinnerSVG } from '@/components/old/Loading'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { Text } from '@/components/old/Typography'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'
import type { TInstallWorkflow } from '@/types'
import { removeSnakeCase, sentanceCase } from '@/utils'

interface IWorkflowApproveAllModal {
  buttonClassName?: string
  buttonVariant?: TButtonVariant
  workflow: TInstallWorkflow
}

export const WorkflowApproveAllModal: FC<IWorkflowApproveAllModal> = ({
  buttonClassName,
  buttonVariant,
  workflow,
}) => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const path = usePathname()

  const workflowId = workflow?.id
  const [isOpen, setIsOpen] = useState<boolean>(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [hasBeenApproved, setHasBeenApproved] = useState(false)

  const { data, error, execute, headers, isLoading } = useServerAction({
    action: approveAllWorkflowSteps,
  })

  const workflowType = removeSnakeCase(workflow?.type)
  const workflowPath = `/${org.id}/installs/${workflow?.owner_id}/workflows/${workflow?.id}`
  const historyPath = `/${org.id}/installs/${workflow?.owner_id}/workflows`

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'workflow_approve_all',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          workflowId,
          workflowType: workflow?.type,
        },
      })
    }

    if (data) {
      setHasBeenApproved(true)
      trackEvent({
        event: 'install_workflow_approve_all',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          workflowId,
          workflowType: workflow?.type,
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
              className="!max-w-xl"
              isOpen={isOpen}
              heading={`Approve pending changes?`}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? (
                  <Notice>
                    {error?.error || 'Unable to approve all workflow steps'}
                  </Notice>
                ) : null}
                <Text>
                  Are you sure you want to approve these changes? This will mark
                  all approval steps as reviewed and allow automatic changes to
                  this install.
                </Text>

                <Text className="mt-3" variant="med-12">
                  Step to approve
                </Text>
                <div className="flex flex-wrap gap-2">
                  {workflow?.steps
                    ?.filter(
                      (s) =>
                        s?.execution_type === 'approval' &&
                        s?.status?.status !== 'discarded' &&
                        !s?.approval?.response
                    )
                    .map((s) => (
                      <Badge className="text-[11px]" variant="code" key={s?.id}>
                        {sentanceCase(s?.name)}
                      </Badge>
                    ))}
                </div>
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
                    execute({
                      body: { approval_option: 'approve-all' },
                      orgId: org.id,
                      workflowId,
                    })
                  }}
                  variant="primary"
                >
                  {isLoading ? (
                    <SpinnerSVG />
                  ) : isKickedOff ? (
                    <CheckIcon size="18" />
                  ) : null}{' '}
                  Approve all
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        disabled={hasBeenApproved}
        className={classNames('text-sm w-fit !h-[32px] !leading-none', {
          [`${buttonClassName}`]: Boolean(buttonClassName),
        })}
        onClick={() => {
          setIsOpen(true)
        }}
        variant={buttonVariant}
      >
        Approve all
      </Button>
    </>
  )
}
