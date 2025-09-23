'use client'

import { useRouter, usePathname } from 'next/navigation'
import { useState } from 'react'
import { createPortal } from 'react-dom'
import { CubeIcon } from '@phosphor-icons/react'
import { createAppInstall } from '@/actions/apps/create-app-install'
import { Button } from '@/components/Button'
import { InstallForm } from '@/components/InstallForm'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
//import { createAppInstall } from './app-actions'
import type { TAppConfig } from '@/types'

interface IAppCreateInstallButton {
  platform: string | 'aws' | 'azure'
}

export const AppCreateInstallButton = ({
  platform,
}: IAppCreateInstallButton) => {
  const [isOpen, setIsOpen] = useState(false)

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
                    Enter the following information to setup your install.
                  </Text>
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
              contentClassName="px-0 py-0"
            >
              <LoadAppConfigs
                platform={platform}
                onClose={() => {
                  setIsOpen(false)
                }}
              />
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="flex items-center gap-2 text-sm font-medium"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <CubeIcon size={16} /> Create install
      </Button>
    </>
  )
}

interface ICreateInstallFromAppConfig {
  onClose: () => void
  platform?: 'aws' | 'azure' | string
}

const LoadAppConfigs = (props: ICreateInstallFromAppConfig) => {
  const { org } = useOrg()
  const { app } = useApp()
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
    <CreateInstallFromAppConfig configId={configs?.at(0)?.id} {...props} />
  )
}

const CreateInstallFromAppConfig = ({
  onClose,
  platform,
  configId,
}: ICreateInstallFromAppConfig & { configId: string }) => {
  const path = usePathname()
  const router = useRouter()
  const { org } = useOrg()
  const { app } = useApp()
  const {
    data: config,
    isLoading,
    error,
  } = useQuery<TAppConfig>({
    path: `/api/orgs/${org?.id}/apps/${app?.id}/configs/${configId}?recurse=true`,
  })

  return (
    <>
      {isLoading ? (
        <div className="p-6">
          <Loading loadingText="Loading configs..." variant="stack" />
        </div>
      ) : error?.error ? (
        <div className="p-6">
          <Notice>{error?.error || 'Unable to load app config.'}</Notice>
        </div>
      ) : (
        <InstallForm
          onSubmit={(formData) => {
            return createAppInstall({
              appId: app.id,
              orgId: org.id,
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
          }}
          onCancel={onClose}
          platform={platform}
          inputConfig={{
            ...config.input,
            input_groups: nestInputsUnderGroups(
              config.input?.input_groups,
              config.input?.inputs
            ),
          }}
        />
      )}
    </>
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
