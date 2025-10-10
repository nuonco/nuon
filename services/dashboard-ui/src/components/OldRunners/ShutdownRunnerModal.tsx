'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { usePathname } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, ArrowClockwiseIcon } from '@phosphor-icons/react'
import { shutdownRunner } from '@/actions/runners/shutdown-runner'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { CheckboxInput } from '@/components/Input'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'

interface IShutdownRunnerModal {
  runnerId: string
}

export const ShutdownRunnerModal: FC<IShutdownRunnerModal> = ({ runnerId }) => {
  const path = usePathname()
  const { user } = useUser()
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [force, setForce] = useState<boolean>(false)

  const {
    data: isShutdown,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({
    action: shutdownRunner,
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
        event: 'runner_shutdown',
        user,
        status: 'error',
        props: {
          orgId: org.id,
          runnerId,
          err: error?.error,
        },
      })
    }

    if (isShutdown) {
      trackEvent({
        event: 'runner_shutdown',
        user,
        status: 'ok',
        props: { orgId: org.id, runnerId },
      })

      setIsOpen(false)
    }
  }, [isShutdown, error, headers])

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
                {error ? (
                  <Notice>
                    {error?.error || 'Unable to shutdown runner.'}
                  </Notice>
                ) : null}
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
                    setIsKickedOff(true)

                    execute({
                      runnerId,
                      orgId: org.id,
                      path,
                      force,
                    })
                  }}
                  variant="primary"
                >
                  {isLoading ? (
                    <SpinnerSVG />
                  ) : isKickedOff ? (
                    <CheckIcon size="18" />
                  ) : (
                    <ArrowClockwiseIcon size="18" />
                  )}{' '}
                  Shutdown runner
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full text-orange-600 dark:text-orange-400"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <ArrowClockwiseIcon size="16" />
        Shutdown runner
      </Button>
    </>
  )
}
