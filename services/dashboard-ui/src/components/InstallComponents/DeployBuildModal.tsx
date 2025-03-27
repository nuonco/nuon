'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import { CloudCheck, CloudArrowUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { RadioInput } from '@/components/Input'
import { SpinnerSVG, Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import { deployComponentBuild } from '@/components/install-actions'
import type { TBuild } from '@/types'
import { trackEvent } from '@/utils'

export const InstallDeployBuildModal: FC<{}> = ({}) => {
  const params =
    useParams<Record<'org-id' | 'install-id' | 'component-id', string>>()
  const orgId = params['org-id']
  const installId = params['install-id']
  const componentId = params['component-id']
  const { user } = useUser()
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

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading={`Deploy build?`}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Select a build to deploy from the list below.
                </Text>

                <BuildOptions
                  componentId={componentId}
                  orgId={orgId}
                  setBuildId={setBuildId}
                />
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

      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        onClick={() => {
          setIsOpen(true)
        }}
        variant="ghost"
      >
        Deploy component build
      </Button>
    </>
  )
}

const BuildOptions: FC<{
  componentId: string
  orgId: string
  setBuildId: (id: string) => void
}> = ({ componentId, orgId, ...props }) => {
  const [builds, setBuilds] = useState<Array<TBuild>>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  useEffect(() => {
    fetch(`/api/${orgId}/components/${componentId}/builds`)
      .then((res) =>
        res.json().then((blds) => {
          setBuilds(
            (blds as Array<TBuild>).filter((b) => b.status === 'active')
          )
          setIsLoading(false)
        })
      )
      .catch((err) => {
        console.error(err?.message)
        setIsLoading(false)
        setError('Unable to load component builds')
      })
  }, [])

  return (
    <div className="w-full max-h-[450px] overflow-y-auto border rounded-md">
      <Text
        className="px-3 py-2 text-cool-grey-600 dark:text-cool-grey-400 border-b"
        variant="med-14"
      >
        Active builds
      </Text>
      {error ? (
        <div className="p-3">
          <Notice>{error}</Notice>
        </div>
      ) : isLoading ? (
        <div className="p-3 text-sm">
          <Loading loadingText="Loading builds..." />
        </div>
      ) : builds && builds?.length ? (
        builds.map((build) => (
          <RadioInput
            className="!items-start"
            key={build?.id}
            name="build-id"
            value={build?.id}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              props.setBuildId(e.target?.value)
            }}
            labelText={
              <span className="flex flex-col gap-0">
                <span className="flex gap-4">
                  <Text variant="med-12">
                    <StatusBadge
                      status={build?.status}
                      isWithoutBorder
                      isStatusTextHidden
                    />
                    {build?.id}
                  </Text>
                  <Time
                    className="!font-normal"
                    variant="reg-12"
                    time={build.created_at}
                  />
                </span>
                {build?.vcs_connection_commit ? (
                  <span>
                    <Text className="!font-normal">
                      <span className="truncate max-w-[50px]">
                        {build?.vcs_connection_commit?.sha}
                      </span>
                      <span>{build?.vcs_connection_commit?.message}</span>
                    </Text>
                  </span>
                ) : null}
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
