'use client'

import { useRouter, useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import { ArrowURightDown, Check } from '@phosphor-icons/react'
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
              className="!max-w-xl"
              isOpen={isOpen}
              heading={`Deprovision entire install`}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-8 mb-12">
                {error ? <Notice>{error}</Notice> : null}
                <span className="flex flex-col gap-1">
                  <Text variant="med-18">
                    Are you sure you want to deprovision {install?.name}?
                  </Text>
                  <Text
                    className="text-cool-grey-600 dark:text-white/70"
                    variant="reg-12"
                  >
                    Deprovisioning an install will remove it from the cloud
                    account.
                  </Text>
                </span>

                <div className="flex flex-col gap-2">
                  <Text variant="reg-14">
                    This will create a workflow that attempts to:
                  </Text>

                  <ul className="flex flex-col gap-1 list-disc pl-4">
                    <li className="text-sm">
                      Teardown each install component according to the
                      dependency order.
                    </li>
                    <li className="text-sm">Teardown the install sandbox</li>
                  </ul>
                </div>

                <div className="w-full">
                  <label className="flex flex-col gap-1 w-full">
                    <Text variant="med-14">
                      To verify, type{' '}
                      <span className="text-red-800 dark:text-red-500">
                        {install?.name}
                      </span>{' '}
                      below.
                    </Text>
                    <Input
                      placeholder="install name"
                      className="w-full"
                      type="text"
                      value={confirm}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        setConfirm(e?.currentTarget?.value)
                      }}
                    />
                  </label>
                </div>
                <div className="flex flex-col items-start">
                  <Text className="!font-normal max-w-sm" variant="reg-12">
                    Sometimes resources can be leaked and prevent deprovision.
                    Would you like to attempt to teardown all components and the
                    sandbox, regardless if a previous step fails?
                  </Text>
                  <CheckboxInput
                    name="ack"
                    defaultChecked={force}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setForceDelete(Boolean(e?.currentTarget?.checked))
                    }}
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px] !items-start"
                    labelText={'Continue deprovision even if steps fail?'}
                  />
                </div>
                <Notice className="max-w-md" variant="warn">
                  Finally, after this has run please manually teardown the
                  CloudFormation stack in the AWS console.
                </Notice>
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
                  disabled={confirm !== install.name}
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    deleteInstall({ installId, orgId, force })
                      .then((workflowId) => {
                        trackEvent({
                          event: 'install_delete',
                          user,
                          status: 'ok',
                          props: { orgId, installId },
                        })

                        if (workflowId) {
                          router.push(
                            `/${orgId}/installs/${installId}/history/${workflowId}`
                          )
                        } else {
                          router.push(`/${orgId}/installs/${installId}/history`)
                        }

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
                    <ArrowURightDown size="18" />
                  )}{' '}
                  Deprovision install
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <ArrowURightDown size="16" /> Deprovision install
      </Button>
    </>
  )
}
