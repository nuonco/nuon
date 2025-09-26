'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { usePathname } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, ArrowsCounterClockwiseIcon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Input } from '@/components/Input'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { updateRunner } from '@/components/runner-actions'
import { trackEvent } from '@/utils'
import type { TRunnerGroupSettings } from '@/types'

interface IUpdateRunnerModal {
  runnerId: string
  settings: TRunnerGroupSettings
  orgId: string
}

export const UpdateRunnerModal: FC<IUpdateRunnerModal> = ({
  runnerId,
  settings,
  orgId,
}) => {
  const pathName = usePathname()
  const { user } = useUser()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [tag, setTag] = useState<string>()
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
                  setIsLoading(true)
                  updateRunner({
                    runnerId,
                    orgId,
                    path: pathName,
                    body: {
                      container_image_tag: tag || '',
                      container_image_url: settings?.container_image_url,
                      org_awsiam_role_arn: settings?.org_aws_iam_role_arn || '',
                      org_k8s_service_account_name:
                        settings?.org_k8s_service_account_name,
                      runner_api_url: settings?.runner_api_url,
                    },
                  }).then((res) => {
                    if (res?.error) {
                      trackEvent({
                        event: 'runner_update',
                        user,
                        status: 'error',
                        props: {
                          orgId,
                          runnerId,
                          err: res.error?.error,
                        },
                      })
                      setError(
                        res?.error?.error ||
                          'Error occured, please refresh page and try again.'
                      )
                      setIsLoading(false)
                    } else {
                      trackEvent({
                        event: 'runner_update',
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
              >
                <div className="flex flex-col gap-4 mb-8">
                  {error ? <Notice>{error}</Notice> : null}
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
                    {isKickedOff ? (
                      <CheckIcon size="18" />
                    ) : isLoading ? (
                      <SpinnerSVG />
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
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3"
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
