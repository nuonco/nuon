'use client'

import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useRouter } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, BoxArrowUpIcon } from '@phosphor-icons/react'
import { reprovisionSandbox } from '@/actions/installs/reprovision-sandbox'
import { Button } from '@/components/Button'
import { CheckboxInput } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/utils'

export const ReprovisionSandboxModal = () => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()

  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [planOnly, setPlanOnly] = useState(false)

  const { data, error, execute, headers, isLoading } = useServerAction({
    action: reprovisionSandbox,
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
        event: 'install_sandbox_reprovision',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: error?.error },
      })
    }

    if (data) {
      trackEvent({
        event: 'install_sandbox_reprovision',
        user,
        status: 'ok',
        props: { orgId: org.id, installId: install.id },
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
  }, [data, error, headers])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading="Reprovision sandbox?"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? (
                  <Notice>
                    {error?.error || 'Unable to kickoff sandbox reprovision'}
                  </Notice>
                ) : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to reprovision this sandbox?
                </Text>

                <CheckboxInput
                  name="ack"
                  defaultChecked={planOnly}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setPlanOnly(Boolean(e?.currentTarget?.checked))
                  }}
                  labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px]"
                  labelText={'Only create a reprovision plan?'}
                />
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
                      body: { plan_only: planOnly },
                      installId: install.id,
                      orgId: org.id,
                    })
                  }}
                  variant="primary"
                >
                  {isLoading ? (
                    <SpinnerSVG />
                  ) : isKickedOff ? (
                    <CheckIcon size="18" />
                  ) : (
                    <BoxArrowUpIcon size="18" />
                  )}{' '}
                  Reprovision sandbox
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
        <BoxArrowUpIcon size="16" />
        Reprovision sandbox
      </Button>
    </>
  )
}
