'use client'

import { useRouter, usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, HammerIcon } from '@phosphor-icons/react'
import { buildComponent } from '@/actions/apps/build-component'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import type { TComponent } from '@/types'
import { trackEvent } from '@/utils'

export const BuildComponentButton = ({
  component,
}: {
  component: TComponent
}) => {
  const path = usePathname()
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { app } = useApp()

  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

  const {
    data: build,
    error,
    isLoading,
    execute,
  } = useServerAction({
    action: buildComponent,
  })

  const handleClose = () => {
    setIsKickedOff(false)
    setIsOpen(false)
  }

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 30000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'component_build',
        user,
        status: 'error',
        props: {
          orgId: org.id,
          appId: app.id,
          componentId: component.id,
        },
      })
    }

    if (build) {
      trackEvent({
        event: 'component_build',
        user,
        status: 'ok',
        props: {
          orgId: org.id,
          appId: app.id,
          componentId: component.id,
        },
      })
      if (build?.id) {
        router.push(`${path}/builds/${build?.id}`)
      }

      setIsOpen(false)
    }
  }, [build, error])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              isOpen={isOpen}
              heading={`Build ${component.name} component?`}
              onClose={handleClose}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error?.error ? <Notice>{error?.error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to build {component.name}?
                </Text>
              </div>
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-base"
                >
                  Cancel
                </Button>
                <Button
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsKickedOff(true)
                    execute({
                      componentId: component.id,
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
                    <HammerIcon size="18" />
                  )}{' '}
                  Build component
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm flex items-center gap-1"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        Build component
      </Button>
    </>
  )
}
