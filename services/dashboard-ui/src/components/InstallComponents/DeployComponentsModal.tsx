'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CloudArrowUpIcon, CloudCheckIcon } from '@phosphor-icons/react'
import { deployComponents } from '@/actions/installs/deploy-components'
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

export const DeployComponentsModal = () => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()
  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [planOnly, setPlanOnly] = useState(false)

  const {
    data: deploysOk,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({ action: deployComponents })

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
        event: 'components_deploy',
        status: 'error',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
          err: error?.error,
        },
      })
    }

    if (deploysOk) {
      trackEvent({
        event: 'components_deploy',
        status: 'ok',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
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
  }, [deploysOk, error, headers])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading="Deploy all components"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? (
                  <Notice>
                    {error?.error || 'Unable to deploy components'}
                  </Notice>
                ) : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to deploy components? This will deploy
                  all components to this install.
                </Text>
                <CheckboxInput
                  name="ack"
                  defaultChecked={planOnly}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setPlanOnly(Boolean(e?.currentTarget?.checked))
                  }}
                  labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px]"
                  labelText={'Plan Only?'}
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
                    <CloudCheckIcon size="18" />
                  ) : (
                    <CloudArrowUpIcon size="18" />
                  )}{' '}
                  Deploy all components
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-fit"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <CloudArrowUpIcon size="16" />
        Deploy all components
      </Button>
    </>
  )
}
