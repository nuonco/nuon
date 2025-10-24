'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { BoxArrowDownIcon, CheckIcon } from '@phosphor-icons/react'
import { deprovisionSandbox } from '@/actions/installs/deprovision-sandbox'
import { Button } from '@/components/old/Button'
import { CheckboxInput, Input } from '@/components/old/Input'
import { SpinnerSVG } from '@/components/old/Loading'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { Text } from '@/components/old/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'

export const DeprovisionSandboxModal = () => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()

  const [confirm, setConfirm] = useState<string>()
  const [force, setForceDelete] = useState(false)
  const [planOnly, setPlanOnly] = useState(false)
  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

  const { data, error, execute, headers, isLoading } = useServerAction({
    action: deprovisionSandbox,
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
        event: 'install_sandbox_deprovision',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: error?.error },
      })
    }

    if (data) {
      trackEvent({
        event: 'install_sandbox_deprovision',
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
              className="!max-w-xl"
              isOpen={isOpen}
              heading={`Deprovision install sandbox`}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-8 mb-12">
                {error ? (
                  <Notice>
                    {error?.error || 'Unable to kickoff sandbox deprovision'}
                  </Notice>
                ) : null}
                <span className="flex flex-col gap-1">
                  <Text variant="med-18">
                    Are you sure you want to deprovision {install?.name}{' '}
                    sandbox?
                  </Text>
                  <Text
                    className="text-cool-grey-600 dark:text-white/70"
                    variant="reg-12"
                  >
                    Deprovisioning a sandbox will remove it from the cloud
                    account.
                  </Text>
                </span>

                <div className="flex flex-col gap-2">
                  <Text variant="reg-14">
                    This will create a workflow that attempts to:
                  </Text>

                  <ul className="flex flex-col gap-1 list-disc pl-4">
                    <li className="text-sm">Teardown the install sandbox</li>
                  </ul>
                </div>

                <div className="w-full">
                  <label className="flex flex-col gap-1 w-full">
                    <Text variant="med-14">
                      To verify, type{' '}
                      <span className="text-red-800 dark:text-red-500">
                        deprovision
                      </span>{' '}
                      below.
                    </Text>
                    <Input
                      placeholder="deprovision"
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
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px]"
                    labelText={'Continue deprovision even if steps fail?'}
                  />
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
                  disabled={confirm !== 'deprovision'}
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsKickedOff(true)
                    execute({
                      body: {
                        plan_only: planOnly,
                        error_behavior: force ? 'continue' : 'abort',
                      },
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
                    <BoxArrowDownIcon size="18" />
                  )}{' '}
                  Deprovision sandbox
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 text-red-800 dark:text-red-500"
        onClick={() => {
          setIsOpen(true)
        }}
        variant="ghost"
      >
        <BoxArrowDownIcon size="16" /> Deprovision sandbox
      </Button>
    </>
  )
}
