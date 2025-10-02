'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, TrashSimpleIcon } from '@phosphor-icons/react'
import { teardownComponent } from '@/actions/installs/teardown-component'
import { Button } from '@/components/Button'
import { CheckboxInput, Input } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'

interface IDeleteComponentModal {
  componentId: string
  componentName: string
}

export const DeleteComponentModal: FC<IDeleteComponentModal> = ({
  componentId,
  componentName,
}) => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()

  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

  const [planOnly, setPlanOnly] = useState(false)
  const [confirm, setConfirm] = useState<string>()
  const [force, setForceDelete] = useState(false)

  const {
    data: teardownOk,
    error,
    execute,
    headers,
    isLoading,
    status,
  } = useServerAction({
    action: teardownComponent,
  })

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'component_delete',
        user,
        status: 'error',
        props: {
          orgId: org.id,
          installId: install.id,
          componentId,
          err: error,
        },
      })
    }

    if (teardownOk) {
      trackEvent({
        event: 'component_delete',
        user,
        status: 'ok',
        props: {
          orgId: org.id,
          installId: install.id,
          componentId,
        },
      })

      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers?.['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
      setIsOpen(false)
    }
  }, [teardownOk, error, headers, status])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-2xl"
              isOpen={isOpen}
              heading={`Teardown ${componentName}`}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-6 mb-12">
                {error ? <Notice>{error?.error}</Notice> : null}
                <span className="flex flex-col gap-1">
                  <Text variant="med-18">
                    Are you sure you want to teardown {componentName}?
                  </Text>
                  <Text variant="reg-12">
                    Tearing down components will affect the working nature of
                    this install.
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
                        teardown
                      </span>{' '}
                      below.
                    </Text>
                    <Input
                      placeholder="teardown"
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
                        <Text variant="med-14">Force teardown</Text>
                        <Text className="!font-normal" variant="reg-12">
                          Force teardown may result in orphaned artifacts that
                          will need manual removal.
                        </Text>
                      </span>
                    }
                  />

                  <CheckboxInput
                    name="ack"
                    defaultChecked={planOnly}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setPlanOnly(Boolean(e?.currentTarget?.checked))
                    }}
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px]"
                    labelText={'Only create a component teardown plan?'}
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
                  disabled={confirm !== 'teardown'}
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsKickedOff(true)
                    execute({
                      body: {
                        plan_only: planOnly,
                        error_behavior: force ? 'continue' : 'abort',
                      },
                      componentId,
                      installId: install.id,
                      orgId: org.id,
                    })
                  }}
                  variant="danger"
                >
                  {isLoading ? (
                    <SpinnerSVG />
                  ) : isKickedOff ? (
                    <CheckIcon size="18" />
                  ) : (
                    <TrashSimpleIcon size="18" />
                  )}{' '}
                  Teardown component
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
        <TrashSimpleIcon size="16" /> Teardown component
      </Button>
    </>
  )
}
