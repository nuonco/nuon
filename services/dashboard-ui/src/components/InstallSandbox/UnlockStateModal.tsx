'use client'

import { usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { LockKeyOpenIcon } from '@phosphor-icons/react'
import { unlockTerraformWorkspace } from '@/actions/runners/unlock-terraform-workspace'
import { Button } from '@/components/Button'
import { Link } from '@/components/Link'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { jobHrefPath, jobName } from '@/components/Runners/helpers'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'

interface IUnlockModal {
  workspace: any
  lock: any
}

export const UnlockModal = ({ workspace, lock }: IUnlockModal) => {
  const path = usePathname()
  const { user } = useUser()
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)

  const {
    data: isUnlocked,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({
    action: unlockTerraformWorkspace,
  })

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'terraform_workspace_state_unlock',
        user,
        status: 'error',
        props: { orgId: org.id, workflowId: workspace.id, err: error?.error },
      })
    }

    if (isUnlocked !== null) {
      trackEvent({
        event: 'terraform_workspace_state_unlock',
        user,
        status: 'ok',
        props: { orgId: org.id, workspaceId: workspace.id },
      })

      setIsOpen(false)
    }
  }, [isUnlocked, error, headers])

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
              {error ? (
                <Notice>
                  {error?.error || 'Unable to unlock Terraform workspace.'}
                </Notice>
              ) : null}
              {lock?.runner_job ? (
                <Text>
                  This Terraform state is associated with this{' '}
                  <Link href={`/${org.id}/${jobHrefPath(lock?.runner_job)}`}>
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
                    execute({
                      orgId: org.id,
                      path,
                      terraformWorkspaceId: workspace.id,
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
                      <LockKeyOpenIcon size="18" />
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
        <LockKeyOpenIcon size="18" />
        Force unlock
      </Button>
    </>
  )
}
