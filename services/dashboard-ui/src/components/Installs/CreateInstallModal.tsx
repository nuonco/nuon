'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Cube, CaretLeft } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { RadioInput } from '@/components/Input'
import { InstallForm } from '@/components/InstallForm'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { createAppInstall } from '../app-actions'
import type { TApp, TAppInputConfig } from '@/types'

interface ICreateInstallModal {}

export const CreateInstallModal: FC<ICreateInstallModal> = ({}) => {
  const [isOpen, setIsOpen] = useState(false)
  const [app, selectApp] = useState<TApp | undefined>()

  const onClose = () => {
    setIsOpen(false)
  }

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-5xl"
              isOpen={isOpen}
              heading={
                <span className="flex flex-col gap-2">
                  <Text variant="med-18">Create install</Text>
                  <Text variant="reg-14" className="!font-normal">
                    Select an app then create an install.
                  </Text>
                </span>
              }
              onClose={onClose}
              contentClassName="px-0 py-0"
            >
              {app ? (
                <CreateInstallFromApp
                  app={app}
                  onClose={onClose}
                  selectApp={selectApp}
                />
              ) : (
                <AppSelect onClose={onClose} selectApp={selectApp} />
              )}
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="flex items-center gap-2 text-sm font-medium w-fit"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <Cube size={16} /> Create install
      </Button>
    </>
  )
}

const AppSelect: FC<{ onClose: () => void; selectApp: (app: TApp) => void }> = (
  props
) => {
  const { org } = useOrg()
  const [isLoading, setIsLoading] = useState(true)
  const [apps, setApps] = useState<Array<TApp> | undefined>()
  const [error, setError] = useState<string>()

  useEffect(() => {
    fetch(`/api/${org?.id}/apps`).then((r) =>
      r.json().then((res) => {
        setIsLoading(false)
        if (res?.error) {
          setError(res?.error?.error || 'Unable to fetch your apps')
        } else {
          setApps(res?.data)
        }
      })
    )
  }, [])

  return isLoading ? (
    <div className="p-6">
      <Loading loadingText="Loading apps..." variant="stack" />
    </div>
  ) : error ? (
    <div className="p-6">
      <Notice>{error}</Notice>
    </div>
  ) : (
    <div>
      {apps?.map((app) =>
        app?.runner_config?.app_runner_type ? (
          <RadioInput
            className="mt-0.5"
            key={app?.id}
            name="app-id"
            value={app?.id}
            onChange={() => {
              props.selectApp(app)
            }}
            labelClassName="!px-6 !items-start"
            labelText={
              <span className="flex flex-col gap-0">
                <span className="flex gap-4">
                  <Text variant="med-12">{app?.name}</Text>
                </span>

                <span>
                  <Text className="!font-normal" isMuted>
                    {app?.id}
                  </Text>
                </span>
              </span>
            }
          />
        ) : null
      )}
      <div className="border-t px-6 py-4 flex justify-end">
        <Button
          className="text-sm font-medium flex items-center"
          onClick={() => {
            props.onClose()
          }}
        >
          Cancel
        </Button>
      </div>
    </div>
  )
}

const CreateInstallFromApp: FC<{
  app: TApp
  onClose: () => void
  selectApp: (app: undefined) => void
}> = ({ app, ...props }) => {
  const { org } = useOrg()
  const [fetchTry, setFetchTry] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [inputConfig, setInputConfig] = useState<TAppInputConfig | undefined>()
  const [error, setError] = useState<string>()
  const router = useRouter()


  const fetchAppInputs = () => {
    setFetchTry(prev => prev++);
    fetch(`/api/${org?.id}/apps/${app?.id}/input-configs/latest`).then((r) =>
      r.json().then((res) => {
        setIsLoading(false)
        if (res?.error) {
          setError('Unable to fetch app input configs')
          if (fetchTry < 5) {
            fetchAppInputs()
          }
        } else {
          setError(undefined)
          setInputConfig(res.data)
        }
      })
    )
}
  
  useEffect(() => {
    fetchAppInputs()
  }, [])

  return (
    <div>
      {isLoading ? (
        <div className="p-6">
          <Loading loadingText="Loading configs..." variant="stack" />
        </div>
      ) : error ? (
        <div className="p-6">
          <Notice>{error}</Notice>
        </div>
      ) : (
        <>
          <div className="px-6 pt-4">
            <Button
              className="text-sm font-medium flex items-center gap-1 !pl-1.5"
              onClick={() => {
                props.selectApp(undefined)
              }}
            >
              <CaretLeft size="16" />
              Back
            </Button>
          </div>
          <InstallForm
            onSubmit={(formData) => {
              return createAppInstall({
                appId: app?.id,
                orgId: org?.id,
                formData,
                platform: app?.runner_config.app_runner_type,
              })
            }}
            onSuccess={(data) => {
              router.push(
                `/${org?.id}/installs/${(data as Record<'installId' | 'workflowId', string>)?.installId}/workflows/${(data as Record<'installId' | 'workflowId', string>)?.workflowId}`
              )
            }}
            onCancel={() => {
              props.onClose()
            }}
            platform={app?.runner_config.app_runner_type}
            inputConfig={inputConfig}
          />
        </>
      )}
    </div>
  )
}
