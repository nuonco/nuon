'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CloudCheckIcon, CloudArrowUpIcon } from '@phosphor-icons/react'
import { deployComponent } from '@/actions/installs/deploy-component'
import { Button, type TButtonVariant } from '@/components/Button'
import { CheckboxInput, RadioInput } from '@/components/Input'
import { SpinnerSVG, Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useServerAction } from '@/hooks/use-server-action'
import type { TBuild } from '@/types'
import { trackEvent } from '@/lib/segment-analytics'

export const InstallDeployBuildModal: FC<{
  buttonClassName?: string
  buttonText?: string
  buttonVariant?: TButtonVariant
  componentId: string
  initBuildId?: string
  initDeployDeps?: boolean
}> = ({
  buttonClassName = 'text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full',
  buttonText = 'Deploy component build',
  buttonVariant = 'ghost',
  initBuildId,
  componentId,
  initDeployDeps = false,
}) => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()

  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [planOnly, setPlanOnly] = useState(false)

  const [buildId, setBuildId] = useState<string>(initBuildId)
  const [deployDeps, setDeployDeps] = useState<boolean>(initDeployDeps)

  const {
    data: deploy,
    error,
    execute,
    isLoading,
    headers,
  } = useServerAction({
    action: deployComponent,
  })

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
        event: 'component_deploy',
        user,
        status: 'error',
        props: {
          orgId: org?.id,
          installId: install.id,
          componentId,
          buildId,
          err: error?.error,
        },
      })
    }

    if (deploy) {
      trackEvent({
        event: 'component_deploy',
        user,
        status: 'ok',
        props: {
          orgId: org?.id,
          installId: install.id,
          componentId,
          buildId,
        },
      })

      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org?.id}/installs/${install.id}/workflows/${headers?.['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org?.id}/installs/${install.id}/workflows`)
      }

      setIsOpen(false)
    }
  }, [deploy, error, headers])

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
                    <Notice>
                      {error?.error ||
                        'Unable to kick off deployment for this component.'}
                    </Notice>
                  </div>
                ) : null}
                <Text variant="reg-14" className="px-6 pt-6 pb-4">
                  Select an active build from the list below and deploy to your
                  install.
                </Text>

                <BuildOptions
                  buildId={buildId}
                  componentId={componentId}
                  setBuildId={setBuildId}
                />
              </div>

              <div className="p-6 border-t">
                <CheckboxInput
                  name="ack"
                  defaultChecked={planOnly}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setPlanOnly(Boolean(e?.currentTarget?.checked))
                  }}
                  labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px]"
                  labelText={'Plan Only?'}
                />
                <div className="flex gap-3 justify-between flex-wrap">
                  <div className="flex items-start">
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
                          <Text variant="med-12">Deploy dependents</Text>
                          <Text className="!font-normal" variant="reg-12">
                            Deploy all dependents as well as the selected build.
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
                      disabled={!buildId || isKickedOff}
                      className="text-base flex items-center gap-1"
                      onClick={() => {
                        setIsKickedOff(true)
                        execute({
                          body: {
                            build_id: buildId,
                            plan_only: planOnly,
                            deploy_dependents: deployDeps,
                          },
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
                      Deploy build
                    </Button>
                  </div>
                </div>
              </div>
            </Modal>,
            document.body
          )
        : null}

      <Button
        className={buttonClassName}
        onClick={() => {
          setIsOpen(true)
        }}
        variant={buttonVariant}
      >
        {buttonText}
      </Button>
    </>
  )
}

const BuildOptions: FC<{
  buildId?: string
  componentId: string
  setBuildId: (id: string) => void
}> = ({ buildId, componentId, ...props }) => {
  const { org } = useOrg()
  const {
    data: builds,
    isLoading,
    error,
  } = useQuery<TBuild[]>({
    path: `/api/orgs/${org.id}/components/${componentId}/builds`,
  })

  useEffect(() => {
    if (!buildId && builds?.length) {
      props.setBuildId(builds?.at(0)?.id)
    }
  }, [builds])

  return (
    <div className="w-full max-h-[450px] overflow-y-auto">
      {error ? (
        <div className="p-6">
          <Notice>{error?.error}</Notice>
        </div>
      ) : isLoading ? (
        <div className="p-6 text-sm">
          <Loading loadingText="Loading builds..." />
        </div>
      ) : builds && builds?.length ? (
        builds.map((build, idx) => (
          <RadioInput
            className="mt-0.5"
            key={build?.id}
            name="build-id"
            value={build?.id}
            defaultChecked={buildId === build?.id || idx === 0}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              props.setBuildId(e.target?.value)
            }}
            labelClassName="!px-6 !items-start"
            labelText={
              <span className="flex flex-col gap-2">
                <span className="flex gap-4">
                  <Text variant="med-12">Build ID: {build?.id}</Text>
                </span>

                {build?.vcs_connection_commit?.message ? (
                  <Text className="!font-normal max-w-[500px]" isMuted>
                    {build?.vcs_connection_commit?.message}
                  </Text>
                ) : null}

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
