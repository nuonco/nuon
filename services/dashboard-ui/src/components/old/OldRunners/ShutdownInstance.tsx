'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { usePathname } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, CloudArrowDownIcon } from '@phosphor-icons/react'
import { shutdownInstance } from '@/actions/runners/shutdown-instance'
import { Button } from '@/components/old/Button'
import { SpinnerSVG } from '@/components/old/Loading'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { Text } from '@/components/old/Typography'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'

interface IShutdownInstanceModal {
  runnerId: string
}

export const ShutdownInstanceModal: FC<IShutdownInstanceModal> = ({ runnerId }) => {
  const path = usePathname()
  const { user } = useUser()
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

  const {
    data: isShutdown,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({
    action: shutdownInstance,
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
        event: 'runner_shutdown_instance',
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
        event: 'runner_shutdown_instance',
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
              heading="Shutdown runner instance?"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-8">
                {error ? (
                  <Notice>
                    {error?.error || 'Unable to shutdown runner instance.'}
                  </Notice>
                ) : null}
                <Text variant="med-18">Shutdown this runner instance.</Text>
                <Text variant="reg-14" className="leading-relaxed max-w-md">
                  The runner VM will be shutdown and restarted.
                </Text>                
                
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
                    })
                  }}
                  variant="primary"
                >
                  {isLoading ? (
                    <SpinnerSVG />
                  ) : isKickedOff ? (
                    <CheckIcon size="18" />
                  ) : (
                    <CloudArrowDownIcon size="18" />
                  )}{' '}
                  Shutdown instance
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
        <CloudArrowDownIcon size="16" />
        Shutdown instance
      </Button>
    </>
  )
}
