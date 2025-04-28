'use client'

import { useParams, useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import { CloudCheck, CloudArrowUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { CheckboxInput, RadioInput } from '@/components/Input'
import { SpinnerSVG, Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
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
  const router = useRouter()
  const { user } = useUser()
  const [isOpen, setIsOpen] = useState(false)
  const [buildId, setBuildId] = useState<string>()
  const [deployDeps, setDeployDeps] = useState<boolean>(false)
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
              className="!max-w-2xl"
              contentClassName="!p-0"
              heading={`Deploy build?`}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col mb-6">
                {error ? (
                  <div className="px-6 pt-6">
                    <Notice>{error}</Notice>
                  </div>
                ) : null}
                <Text variant="reg-14" className="px-6 pt-6 pb-4">
                  Select an active build from the list below and deploy to your
                  install.
                </Text>

                <BuildOptions
                  componentId={componentId}
                  orgId={orgId}
                  setBuildId={setBuildId}
                />
              </div>

              <div className="flex gap-3 justify-between border-t p-6 flex-wrap">
                <div className="flex items-start px-6">
                  <CheckboxInput
                    name="ack"
                    defaultChecked={deployDeps}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setDeployDeps(Boolean(e?.currentTarget?.checked))
                    }}
                    className="mt-1.5"
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[250px] !items-start"
                    labelText={
                      <span className="flex flex-col gap-1">
                        <Text variant="med-12">Deploy dependencies</Text>
                        <Text className="!font-normal" variant="reg-12">
                          Deploy all dependencies as well as the selected build.
                        </Text>
                      </span>
                    }
                  />
                </div>
                <div className="flex gap-3 items-center">
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
                      deployComponentBuild({
                        buildId,
                        installId,
                        orgId,
                        deployDeps,
                      })
                        .then((workflowId) => {
                          trackEvent({
                            event: 'component_deploy',
                            user,
                            status: 'ok',
                            props: { orgId, installId, componentId, buildId },
                          })
                          setIsLoading(false)
                          setIsKickedOff(true)

                          if (workflowId) {
                            router.push(
                              `/${orgId}/installs/${installId}/history/${workflowId}`
                            )
                          } else {
                            router.push(
                              `/${orgId}/installs/${installId}/history`
                            )
                          }

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
    <div className="w-full max-h-[450px] overflow-y-auto">
      {error ? (
        <div className="p-6">
          <Notice>{error}</Notice>
        </div>
      ) : isLoading ? (
        <div className="p-6 text-sm">
          <Loading loadingText="Loading builds..." />
        </div>
      ) : builds && builds?.length ? (
        builds.map((build) => (
          <RadioInput
            className="mt-0.5"
            key={build?.id}
            name="build-id"
            value={build?.id}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              props.setBuildId(e.target?.value)
            }}
            labelClassName="!px-6 !items-start"
            labelText={
              <span className="flex flex-col gap-0">
                <span className="flex gap-4">
                  <Text variant="med-12">Build ID: {build?.id}</Text>
                </span>

                <span>
                  <Text className="!font-normal text-cool-grey-600 dark:text-white/70">
                    {build?.created_by?.email} created on{' '}
                    <Time time={build?.updated_at} format="long" />
                  </Text>
                </span>
              </span>
            }
          />
        ))
      ) : (
        <Text className="text-sm px-6 pb-2">No active builds found</Text>
      )}
    </div>
  )
}
