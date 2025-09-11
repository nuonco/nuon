'use client'

import classNames from 'classnames'
import { useParams, usePathname, useRouter } from 'next/navigation'
import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { Check, XSquare } from '@phosphor-icons/react'
import { Badge } from '@/components/Badge'
import { Button, type TButtonVariant } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { installWorkflowApproveAll } from '@/components/workflow-actions'
import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflow } from '@/types'
import { trackEvent, removeSnakeCase, sentanceCase } from '@/utils'

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
  const { user } = useUser()
  const { org } = useOrg()
  const pathName = usePathname()
  const params =
    useParams<Record<'org-id' | 'install-id' | 'workflow-id', string>>()
  const router = useRouter()
  const orgId = params?.['org-id']
  const workflowId = workflow?.id
  const [isOpen, setIsOpen] = useState<boolean>(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [hasBeenApproved, setHasBeenApproved] = useState(false)
  const [error, setError] = useState<string>()

  const workflowType = removeSnakeCase(workflow?.type)
  const workflowPath = `/${orgId}/installs/${workflow?.owner_id}/workflows/${workflow?.id}`
  const historyPath = `/${orgId}/installs/${workflow?.owner_id}/workflows`

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
                {error ? <Notice>{error}</Notice> : null}
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
                    ?.filter((s) => s?.execution_type === 'approval' && s?.status?.status !== 'discarded' )
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
                    setIsLoading(true)
                    installWorkflowApproveAll({
                      orgId,
                      workflowId,
                    }).then((res) => {
                      if (res?.error) {
                        trackEvent({
                          event: 'install_workflow_approve_all',
                          status: 'error',
                          user,
                          props: {
                            workflowType: workflow?.type,
                            orgId: org.id,
                            workflowId,
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
                          event: 'install_workflow_approve_all',
                          status: 'ok',
                          user,
                          props: {
                            workflowType: workflow?.type,
                            orgId: org.id,
                            workflowId,
                          },
                        })
                        setIsLoading(false)
                        setIsKickedOff(true)
                        if (
                          pathName !== workflowPath &&
                          pathName !== historyPath
                        ) {
                          router.push(workflowPath)
                        }

                        setIsOpen(false)
                        setHasBeenApproved(true)
                      }
                    })
                  }}
                  variant="primary"
                >
                  {isKickedOff ? (
                    <Check size="18" />
                  ) : isLoading ? (
                    <SpinnerSVG />
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
        className={classNames('text-sm !font-medium w-fit', {
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
