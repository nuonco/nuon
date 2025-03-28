'use client'

import { useRouter, useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Check, TrashSimple } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { CheckboxInput, Input } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { deleteInstall } from '@/components/install-actions'
import { trackEvent } from '@/utils'
import { TInstall } from '@/types'

interface IDeleteInstallModal {
  install: TInstall
}

export const DeleteInstallModal: FC<IDeleteInstallModal> = ({ install }) => {
  const router = useRouter()
  const params = useParams<Record<'org-id' | 'install-id', string>>()

  const installId = params?.['install-id']
  const orgId = params?.['org-id']

  const { user } = useUser()
  const [confirm, setConfirm] = useState<string>()
  const [force, setForceDelete] = useState(false)
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState<string>()

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-2xl"
              isOpen={isOpen}
              heading={`Delete install`}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <span>
                  <Text variant="med-14">
                    Are you sure you want to delete {install?.name}?
                  </Text>
                  <Text variant="reg-14">
                    Deleteing install will remove it from the cloud account.
                  </Text>
                </span>
                <Notice>
                  Warning, this action is not reversible. Please be certain.
                </Notice>

                <div className="w-full">
                  <label className="flex flex-col gap-1 w-full">
                    <Text variant="med-14">
                      To verify, type{' '}
                      <span className="text-red-800 dark:text-red-500">
                        delete
                      </span>{' '}
                      below.
                    </Text>
                    <Input
                      placeholder="delete"
                      className="w-full"
                      type="text"
                      value={confirm}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        setConfirm(e?.currentTarget?.value)
                      }}
                    />
                  </label>
                </div>
                <div className="flex items-start">
                  <CheckboxInput
                    name="ack"
                    defaultChecked={force}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setForceDelete(Boolean(e?.currentTarget?.checked))
                    }}
                    className="mt-1.5"
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px] !items-start"
                    labelText={
                      <span className="flex flex-col gap2">
                        <Text variant="med-14">Force delete</Text>
                        <Text className="!font-normal" variant="reg-12">
                          Force deleting may result in orphaned artifacts that
                          will need manual removal.
                        </Text>
                      </span>
                    }
                  />
                </div>
              </div>
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-sm"
                >
                  Cancel
                </Button>
                <Button
                  disabled={confirm !== 'delete'}
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    deleteInstall({ installId, orgId, force })
                      .then(() => {
                        trackEvent({
                          event: 'install_delete',
                          user,
                          status: 'ok',
                          props: { orgId, installId },
                        })
                        router.push(`/${orgId}/installs`)
                        setForceDelete(false)
                        setIsLoading(false)
                        setIsKickedOff(true)
                        setIsOpen(false)
                      })
                      .catch((err) => {
                        trackEvent({
                          event: 'install_delete',
                          user,
                          status: 'error',
                          props: { orgId, installId, err },
                        })
                        setError(
                          'Error occured, please refresh page and try again.'
                        )
                        setIsLoading(false)
                      })
                  }}
                  variant="danger"
                >
                  {isKickedOff ? (
                    <Check size="18" />
                  ) : isLoading ? (
                    <SpinnerSVG />
                  ) : (
                    <TrashSimple size="18" />
                  )}{' '}
                  Delete install
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full text-red-800 dark:text-red-500"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <TrashSimple size="16" /> Delete install
      </Button>
    </>
  )
}
