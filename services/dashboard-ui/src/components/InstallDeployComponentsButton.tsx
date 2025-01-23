'use client'

import React, { type FC, useEffect, useState } from 'react'
import { CloudCheck, CloudArrowUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { RadioInput } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import {
  deployComponents,
  deployComponentBuild,
} from '@/components/install-actions'
import type { TBuild } from '@/types'

export const InstallDeployComponentButton: FC<{
  installId: string
  orgId: string
  onComplete: () => void
}> = ({ installId, orgId, ...props }) => {
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

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
    <Button
      className="text-base flex items-center gap-1"
      onClick={() => {
        setIsLoading(true)
        deployComponents({ installId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
          props.onComplete()
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
      Deploy components
    </Button>
  )
}

export const InstallDeployLatestBuildButton: FC<{
  builds: Array<TBuild>
  installId: string
  orgId: string
}> = ({ builds, installId, orgId }) => {
  const [isDeploymentOpen, setIsDeploymentOpen] = useState(false)
  const [buildId, setBuildId] = useState<string>()
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

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
      <Modal
        className="max-w-lg"
        heading={`Deploy build ${buildId}?`}
        isOpen={isDeploymentOpen}
        onClose={() => {
          setIsDeploymentOpen(false)
        }}
      >
        <div className="mb-6">
          <Text variant="reg-14" className="leading-relaxed">
            Are you sure you want to deploy build {buildId}? This will replace
            the current install component with the selected build.
          </Text>
        </div>
        <div className="flex gap-3 justify-end">
          <Button
            onClick={() => {
              setIsDeploymentOpen(false)
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
              deployComponentBuild({ buildId, installId, orgId }).then(() => {
                setIsLoading(false)
                setIsKickedOff(true)
                setIsDeploymentOpen(false)
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
      </Modal>
      <Dropdown
        alignment="right"
        className="text-sm !font-medium !p-2 h-[32px]"
        id="deploy-build"
        text="Deploy build"
        isDownIcon
      >
        <div className="min-w-[180px] rounded-md overflow-hidden">
          <div className="w-full max-h-[250px] overflow-y-auto">
            <Text className="px-3 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
              Recent builds
            </Text>
            {builds.map((build) =>
              build.status === 'active' ? (
                <RadioInput
                  key={build?.id}
                  name="build-id"
                  value={build?.id}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setBuildId(e.target?.value)
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
              ) : null
            )}
          </div>
          <hr />
          <Button
            disabled={!buildId}
            className="w-full !rounded-t-none !text-sm flex items-center justify-center gap-2 pl-4"
            onClick={() => {
              setIsDeploymentOpen(true)
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
