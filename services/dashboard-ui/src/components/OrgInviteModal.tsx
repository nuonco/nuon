'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { UserPlus } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Input } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { inviteUserToOrg } from '@/components/org-actions'
import { useOrg } from '@/hooks/use-org'

interface IOrgInviteModal {}

export const OrgInviteModal: FC<IOrgInviteModal> = ({}) => {
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()

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
                  setIsLoading(true)
                  const formData = new FormData(e.currentTarget)

                  inviteUserToOrg(formData, org.id)
                    .then(() => {
                      setIsLoading(false)
                      setIsOpen(false)
                    })
                    .catch((err) => {
                      console.error(err)
                      setIsLoading(false)
                      setError(
                        'Unable to invite user, refresh page and try again.'
                      )
                    })
                }}
              >
                <div className="p-6 flex flex-col gap-4">
                  {error ? <Notice>{error}</Notice> : null}
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
                      setError(undefined)
                      setIsLoading(false)
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
                        <UserPlus size="16" /> Invite user
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
        <UserPlus size="16" />
        Invite team member
      </Button>
    </>
  )
}
