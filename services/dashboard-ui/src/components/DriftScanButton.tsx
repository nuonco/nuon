'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CubeFocusIcon } from '@phosphor-icons/react'
import { deployComponent } from '@/actions/installs/deploy-component'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { trackEvent } from '@/lib/segment-analytics'

import { BuildOptions } from '@/components/InstallComponents/DeployBuildModal'

export const DriftScanButton = ({
  componentId,
  initBuildId,
}: {
  componentId: string
  initBuildId: string
}) => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()

  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [buildId, setBuildId] = useState<string>(initBuildId)

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
        event: 'component_drift_scan',
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
        event: 'component_drift_scan',
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
              heading={`Drift scan build?`}
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
                        'Unable to kick off drift scan for this component.'}
                    </Notice>
                  </div>
                ) : null}
                <Text variant="reg-14" className="px-6 pt-6 pb-4">
                  Select an active build from the list below to preform a drift
                  scan.
                </Text>

                <BuildOptions
                  buildId={buildId}
                  componentId={componentId}
                  setBuildId={setBuildId}
                />
              </div>

              <div className="p-6 border-t">
                <div className="flex gap-3 justify-between flex-wrap">
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
                      className="text-sm flex items-center gap-1"
                      onClick={() => {
                        setIsKickedOff(true)
                        execute({
                          body: {
                            build_id: buildId,
                            plan_only: true,
                            deploy_dependents: false,
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
                        <CubeFocusIcon size="18" />
                      ) : (
                        <CubeFocusIcon size="18" />
                      )}{' '}
                      Drift scan
                    </Button>
                  </div>
                </div>
              </div>
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
        <CubeFocusIcon size="18" />
        Drift Scan
      </Button>
    </>
  )
}
