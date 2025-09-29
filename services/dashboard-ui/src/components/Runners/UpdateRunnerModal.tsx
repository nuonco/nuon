'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { usePathname } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, ArrowsCounterClockwiseIcon } from '@phosphor-icons/react'
import { updateRunner } from '@/actions/runners/update-runner'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Input } from '@/components/Input'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/utils'
import type { TRunnerGroupSettings } from '@/types'

interface IUpdateRunnerModal {
  runnerId: string
  settings: TRunnerGroupSettings
}

export const UpdateRunnerModal: FC<IUpdateRunnerModal> = ({
  runnerId,
  settings,
}) => {
  const path = usePathname()
  const { user } = useUser()
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [tag, setTag] = useState<string>()

  const {
    data: isUpdated,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({
    action: updateRunner,
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
        event: 'runner_update',
        user,
        status: 'error',
        props: {
          orgId: org.id,
          runnerId,
          err: error?.error,
        },
      })
    }

    if (isUpdated) {
      trackEvent({
        event: 'runner_update',
        user,
        status: 'ok',
        props: { orgId: org.id, runnerId },
      })

      setIsOpen(false)
    }
  }, [isUpdated, error, headers])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-xl"
              heading={
                <>
                  <ArrowsCounterClockwiseIcon />
                  Update runner
                </>
              }
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <form
                onSubmit={(e) => {
                  e.preventDefault()
                  setIsKickedOff(true)
                  execute({
                    runnerId,
                    orgId: org.id,
                    path,
                    body: {
                      container_image_tag: tag || '',
                      container_image_url: settings?.container_image_url,
                      org_awsiam_role_arn: settings?.org_aws_iam_role_arn || '',
                      org_k8s_service_account_name:
                        settings?.org_k8s_service_account_name,
                      runner_api_url: settings?.runner_api_url,
                    },
                  })
                }}
              >
                <div className="flex flex-col gap-4 mb-8">
                  {error ? (
                    <Notice>
                      {error?.error || 'Unable to update runner.'}
                    </Notice>
                  ) : null}
                  <Text variant="med-18">
                    Update to a different runner version.
                  </Text>

                  <label className="flex flex-col gap-2">
                    <Text variant="med-14">
                      Enter the runner tag you&apos;d like to update to.
                    </Text>
                    <Input
                      required
                      onChange={(e) => {
                        setTag(e?.currentTarget?.value)
                      }}
                      placeholder="runner tag"
                    />
                  </label>
                </div>
                <div className="flex gap-3 justify-end">
                  <Button
                    type="reset"
                    onClick={() => {
                      setIsOpen(false)
                    }}
                    className="text-sm"
                  >
                    Cancel
                  </Button>
                  <Button
                    className="text-sm flex items-center gap-1"
                    variant="primary"
                  >
                    {isLoading ? (
                      <SpinnerSVG />
                    ) : isKickedOff ? (
                      <CheckIcon size="18" />
                    ) : (
                      <ArrowsCounterClockwiseIcon size="18" />
                    )}{' '}
                    Update runner
                  </Button>
                </div>
              </form>
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
        <ArrowsCounterClockwiseIcon size="16" />
        Update runner
      </Button>
    </>
  )
}
