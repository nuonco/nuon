'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CloudCheck, CloudArrowUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { RadioInput } from '@/components/Input'
import { SpinnerSVG, Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { useOrg } from '@/components/Orgs'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import { deployComponentBuild } from '@/components/install-actions'
import { useQuery } from '@/hooks/use-query'
import type { TBuild } from '@/types'
import { trackEvent } from '@/utils'

export const InstallDeployLatestBuildButton: FC<{
  componentId: string
  installId: string
  orgId: string
}> = ({ componentId, installId, orgId }) => {
  const { user } = useUser()
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [buildId, setBuildId] = useState<string>()
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
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

  return org?.features?.['install-delete-components'] ? null : (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading={`Deploy build ${buildId}?`}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to deploy build {buildId}? This will
                  replace the current install component with the selected build.
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
                  disabled={!buildId}
                  className="text-base flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    deployComponentBuild({ buildId, installId, orgId })
                      .then(() => {
                        trackEvent({
                          event: 'component_deploy',
                          user,
                          status: 'ok',
                          props: { orgId, installId, componentId, buildId },
                        })
                        setIsLoading(false)
                        setIsKickedOff(true)
                        setIsOpen(false)
                      })
                      .catch((err) => {
                        trackEvent({
                          event: 'component_deploy',
                          user,
                          status: 'error',
                          props: {
                            orgId,
                            installId,
                            componentId,
                            buildId,
                            err,
                          },
                        })
                        console.error(err?.message)
                        setIsLoading(false)
                        setError('Unable to create deployment.')
                      })
                  }}
                  variant="primary"
                >
                  {isKickedOff ? (
                    <CloudCheck size="18" />
                  ) : isLoading ? (
                    <SpinnerSVG />
                  ) : (
                    <CloudArrowUp size="18" />
                  )}{' '}
                  Deploy build
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Dropdown
        alignment="right"
        className="text-sm !font-medium !p-2 h-[32px]"
        id="deploy-build"
        text="Deploy build"
        isDownIcon
        wrapperClassName="z-20"
      >
        <div className="min-w-[180px] rounded-md overflow-hidden">
          <BuildOptions
            componentId={componentId}
            orgId={orgId}
            setBuildId={setBuildId}
          />
          <hr />
          <Button
            disabled={!buildId}
            className="w-full !rounded-t-none !text-sm flex items-center justify-center gap-2 pl-4"
            onClick={() => {
              setIsOpen(true)
            }}
            variant="ghost"
          >
            Confirm deploy
          </Button>
        </div>
      </Dropdown>
    </>
  )
}

const BuildOptions: FC<{
  componentId: string
  orgId: string
  setBuildId: (id: string) => void
}> = ({ componentId, orgId, ...props }) => {
  const {
    data: builds,
    isLoading,
    error,
  } = useQuery<TBuild[]>({
    path: `/api/${orgId}/components/${componentId}/builds`,
  })

  return (
    <div className="w-full max-h-[250px] overflow-y-auto">
      <Text className="px-3 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
        Recent builds
      </Text>
      {error ? (
        <div className="p-3">
          <Notice>{error?.error}</Notice>
        </div>
      ) : isLoading ? (
        <div className="p-3 text-sm">
          <Loading loadingText="Loading builds..." />
        </div>
      ) : builds && builds?.length ? (
        builds.map((build) => (
          <RadioInput
            key={build?.id}
            name="build-id"
            value={build?.id}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              props.setBuildId(e.target?.value)
            }}
            labelText={
              <span>
                <Text variant="med-12">{build?.id}</Text>
                <Time
                  className="!font-normal"
                  variant="reg-12"
                  time={build.created_at}
                />
              </span>
            }
          />
        ))
      ) : (
        <Text className="text-sm px-3 pb-2">No active builds found</Text>
      )}
    </div>
  )
}
