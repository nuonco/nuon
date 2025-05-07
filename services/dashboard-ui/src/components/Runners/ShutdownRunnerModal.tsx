'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { usePathname } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Check, ArrowClockwise } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { CheckboxInput } from '@/components/Input'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { shutdownRunner } from '@/components/runner-actions'
import { trackEvent } from '@/utils'

interface IShutdownRunnerModal {
  runnerId: string
  orgId: string
}

export const ShutdownRunnerModal: FC<IShutdownRunnerModal> = ({
  runnerId,
  orgId,
}) => {
  const pathName = usePathname()
  const { user } = useUser()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [force, setForce] = useState<boolean>(false)
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
              heading="Shutdown runner?"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-8">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="med-18">Shutdown this runner gracefully.</Text>
                <Text variant="reg-14" className="leading-relaxed max-w-md">
                  The runner will make a best effort to shut down after any
                  queued jobs are complete.
                </Text>

                <ul className="flex flex-col gap-1 list-disc pl-4">
                  <li className="text-sm">
                    Causes all jobs to queue while the runner restarts
                  </li>
                  <li className="text-sm">
                    Any new version updates will be applied
                  </li>
                  <li className="text-sm">All local state will be refreshed</li>
                </ul>

                <div className="flex items-start">
                  <CheckboxInput
                    name="ack"
                    defaultChecked={force}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setForce(Boolean(e?.currentTarget?.checked))
                    }}
                    className="mt-1.5"
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-sm !items-start"
                    labelText={
                      <span className="flex flex-col gap-1">
                        <Text variant="med-12">Force shutdown</Text>
                        <Text className="!font-normal" variant="reg-12">
                          Immediately shutdown the runner, terminating any
                          in-flight jobs. This has the potential for loss of
                          state.
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
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    shutdownRunner({
                      runnerId,
                      orgId,
                      path: pathName,
                      force,
                    }).then((res) => {
                      if (res?.error) {
                        trackEvent({
                          event: 'runner_shutdown',
                          user,
                          status: 'error',
                          props: { orgId, runnerId, err: res.error?.error },
                        })
                        setError(
                          res?.error?.error ||
                            'Error occured, please refresh page and try again.'
                        )
                        setIsLoading(false)
                      } else {
                        trackEvent({
                          event: 'runner_shutdown',
                          user,
                          status: 'ok',
                          props: { orgId, runnerId },
                        })
                        setIsLoading(false)
                        setIsKickedOff(true)
                        setIsOpen(false)
                      }
                    })
                  }}
                  variant="primary"
                >
                  {isKickedOff ? (
                    <Check size="18" />
                  ) : isLoading ? (
                    <SpinnerSVG />
                  ) : (
                    <ArrowClockwise size="18" />
                  )}{' '}
                  Shutdown runner
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <ArrowClockwise size="16" />
        Shutdown runner
      </Button>
    </>
  )
}
