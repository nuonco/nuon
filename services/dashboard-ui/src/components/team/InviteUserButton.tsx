'use client'

import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { UserPlus } from '@phosphor-icons/react'
import { inviteUser } from '@/actions/orgs/invite-user'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'

export const InviteUserButton = () => {
  const [isOpen, setIsOpen] = useState(false)
  const { org } = useOrg()

  const {
    data: invite,
    error,
    execute,
    isLoading,
  } = useServerAction({
    action: inviteUser,
  })

  useEffect(() => {
    if (invite) {
      setIsOpen(false)
    }
  }, [invite])

  const handleClose = () => {
    setIsOpen(false)
  }

  return (
    <>
      <Button
        variant="secondary"
        onClick={() => setIsOpen(true)}
      >
        <UserPlus size={16} weight="bold" />
        Invite user
      </Button>

      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-lg"
              contentClassName="!p-0"
              heading="Invite team member"
              isOpen={isOpen}
              onClose={handleClose}
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
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="invite-email">
                      Email address of the user you want to invite
                    </Label>
                    <Input
                      id="invite-email"
                      placeholder="user@email.com"
                      type="email"
                      name="email"
                      required
                    />
                  </div>
                </div>
                <div className="p-6 border-t flex gap-3 justify-end">
                  <Button
                    variant="secondary"
                    onClick={handleClose}
                    type="button"
                  >
                    Cancel
                  </Button>
                  <Button variant="primary" type="submit" disabled={isLoading}>
                    <UserPlus size={16} />
                    {isLoading ? 'Inviting...' : 'Invite user'}
                  </Button>
                </div>
              </form>
            </Modal>,
            document.body
          )
        : null}
    </>
  )
}

