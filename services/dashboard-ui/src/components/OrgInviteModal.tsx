'use client'

import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { UserPlusIcon } from '@phosphor-icons/react'
import { inviteUser } from '@/actions/orgs/invite-user'
import { Button } from '@/components/Button'
import { Input } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'

export const OrgInviteModal = () => {
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)

  const {
    data: invite,
    error,
    execute,
    isLoading,
  } = useServerAction({
    action: inviteUser,
  })

  useEffect(() => {
    if (error) {
    }

    if (invite) {
      setIsOpen(false)
    }
  }, [invite, error])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-2xl"
              contentClassName="!p-0"
              heading="Invite team member"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <form
                onSubmit={(e: React.FormEvent<HTMLFormElement>) => {
                  e.preventDefault()
                  const formData = Object.fromEntries(
                    new FormData(e.currentTarget)
                  ) as { email: string }

                  execute({ body: { email: formData?.email }, orgId: org.id })
                }}
              >
                <div className="p-6 flex flex-col gap-4">
                  {error ? (
                    <Notice>
                      {error?.error || 'Unable to invite user to organization.'}
                    </Notice>
                  ) : null}
                  <label className="w-full flex flex-col gap-2">
                    <Text variant="med-14">
                      Email address of the user you want to invite
                    </Text>
                    <Input
                      placeholder="user@email.com"
                      type="email"
                      name="email"
                      required
                    />
                  </label>
                </div>
                <div className="p-6 border-t flex gap-3 justify-end">
                  <Button
                    className="text-sm"
                    onClick={() => {
                      setIsOpen(false)
                    }}
                    type="button"
                  >
                    Cancel
                  </Button>
                  <Button
                    className="text-sm flex items-center gap-2 font-medium"
                    disabled={isLoading}
                    variant="primary"
                  >
                    {isLoading ? (
                      <>
                        <SpinnerSVG /> Inviting...
                      </>
                    ) : (
                      <>
                        <UserPlusIcon size="16" /> Invite user
                      </>
                    )}
                  </Button>
                </div>
              </form>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm flex items-center gap-2"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <UserPlusIcon size="16" />
        Invite team member
      </Button>
    </>
  )
}
