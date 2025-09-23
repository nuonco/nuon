'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useState } from 'react'
import { createPortal } from 'react-dom'
import { CubeIcon, CaretLeftIcon } from '@phosphor-icons/react'
import { createAppInstall } from '@/actions/apps/create-app-install'
import { Button } from '@/components/Button'
import { RadioInput } from '@/components/Input'
import { InstallForm } from '@/components/InstallForm'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TApp, TAppConfig } from '@/types'

interface ICreateInstallModal {}

export const CreateInstallModal = ({}: ICreateInstallModal) => {
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
                <LoadAppConfigs
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
        <CubeIcon size={16} /> Create install
      </Button>
    </>
  )
}

const AppSelect = (props: {
  onClose: () => void
  selectApp: (app: TApp) => void
}) => {
  const { org } = useOrg()
  const {
    data: apps,
    isLoading,
    error,
  } = useQuery<TApp[]>({
    path: `/api/orgs/${org.id}/apps`,
  })

  return isLoading ? (
    <div className="p-6">
      <Loading loadingText="Loading apps..." variant="stack" />
    </div>
  ) : error?.error ? (
    <div className="p-6">
      <Notice>{error?.error || 'Unable to load apps'}</Notice>
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

interface ICreateInstallFromApp {
  app: TApp
  onClose: () => void
  selectApp: (app: undefined) => void
}

const LoadAppConfigs = ({ app, ...props }: ICreateInstallFromApp) => {
  const { org } = useOrg()
  const {
    data: configs,
    isLoading,
    error,
  } = useQuery<TAppConfig[]>({
    path: `/api/orgs/${org?.id}/apps/${app?.id}/configs`,
  })
  return isLoading ? (
    <div className="p-6">
      <Loading loadingText="Loading configs..." variant="stack" />
    </div>
  ) : error?.error ? (
    <div className="p-6">
      <Notice>{error?.error || 'Unable to load app configs'}</Notice>
    </div>
  ) : (
    <CreateInstallFromApp app={app} configId={configs?.at(0)?.id} {...props} />
  )
}

const CreateInstallFromApp = ({
  app,
  configId,
  ...props
}: ICreateInstallFromApp & { configId: string }) => {
  const path = usePathname()
  const router = useRouter()
  const { org } = useOrg()
  const {
    data: config,
    isLoading,
    error,
  } = useQuery<TAppConfig>({
    path: `/api/orgs/${org?.id}/apps/${app?.id}/configs/${configId}?recurse=true`,
  })

  return (
    <div>
      {isLoading ? (
        <div className="p-6">
          <Loading loadingText="Loading configs..." variant="stack" />
        </div>
      ) : error?.error ? (
        <div className="p-6">
          <Notice>{error?.error}</Notice>
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
              <CaretLeftIcon size="16" />
              Back
            </Button>
          </div>
          <InstallForm
            onSubmit={(formData) => {
              return createAppInstall({
                appId: app?.id,
                orgId: org?.id,
                formData,
                path,
              })
            }}
            onSuccess={({ data: install, error, headers, status }) => {
              if (!error && status === 201) {
                router.push(
                  `/${org.id}/installs/${install?.id}/workflows/${headers?.['x-nuon-install-workflow-id']}`
                )
              }

              props.onClose()
            }}
            onCancel={() => {
              props.onClose()
            }}
            platform={app?.runner_config.app_runner_type}
            inputConfig={{
              ...config.input,
              input_groups: nestInputsUnderGroups(
                config.input?.input_groups,
                config.input?.inputs
              ),
            }}
          />
        </>
      )}
    </div>
  )
}

function nestInputsUnderGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups.map((group) => ({
    ...group,
    app_inputs: inputs.filter((input) => input.group_id === group.id),
  }))
}
