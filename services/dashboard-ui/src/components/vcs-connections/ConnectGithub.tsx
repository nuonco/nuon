'use client'

import { usePathname } from 'next/navigation'
import React, { useRef, useState, type FormEvent } from 'react'
import { createVCSConnection } from '@/actions/vcs-connection/create-vcs-connection'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { GITHUB_APP_NAME } from '@/configs/github-app'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'

interface IConnectGithubModal extends IModal {}

export const ConnectGithubModal = (props: IConnectGithubModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const [isManualMode, setIsManualMode] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)

  const { data, error, isLoading, execute } = useServerAction({
    action: createVCSConnection,
  })

  useServerActionToast({
    data,
    error,
    successHeading: 'GitHub connected',
    successContent: <Text>Your GitHub connection has been added.</Text>,
    errorHeading: 'Failed to connect GitHub',
    errorContent: (
      <>
        <Text>Unable to create VCS connection.</Text>
        <Text variant="subtext">{error?.error}</Text>
      </>
    ),
    onSuccess: () => {
      removeModal(props.modalId)
      setIsManualMode(false)
    },
  })

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const formData = Object.fromEntries(new FormData(e.currentTarget))

    execute({
      body: {
        github_install_id: formData.github_install_id as string,
      },
      orgId: org.id,
      path,
    })
  }

  const modalProps = isManualMode
    ? {
        primaryActionTrigger: {
          children: isLoading ? (
            <span className="flex items-center gap-2">
              <Icon variant="Loading" />
              Adding GitHub connection...
            </span>
          ) : (
            <span className="flex items-center gap-2">
              <Icon variant="Plus" />
              Add GitHub connection
            </span>
          ),
          onClick: () => formRef.current?.requestSubmit(),
          disabled: isLoading,
          variant: 'primary' as const,
        },
      }
    : {}

  return (
    <Modal
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="GitHub" />
          <Text variant="h3" weight="strong">
            Connect GitHub to Nuon
          </Text>
        </div>
      }
      {...modalProps}
      {...props}
    >
      {!isManualMode ? (
        <div className="flex flex-col gap-6">
          <Button
            href={`https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${org.id}`}
            variant="ghost"
            className="flex flex-col items-center justify-center gap-4 !p-8 rounded !h-auto !text-center !border-cool-grey-400 dark:!border-dark-grey-500"
          >
            <Text variant="base" weight="strong">
              New GitHub connection
            </Text>
            <Text
              variant="body"
              className="!inline-block text-balance !text-center leading-relaxed"
            >
              Add a new GitHub connection to your Nuon org by installing the{' '}
              <Badge className="!inline-block" variant="code" size="md">
                {GITHUB_APP_NAME}
              </Badge>{' '}
              GitHub app and allowing access to the repositories of your choice.
            </Text>
          </Button>

          <Divider dividerWord="OR" />

          <Button
            onClick={() => setIsManualMode(true)}
            variant="ghost"
            className="flex flex-col items-center justify-center gap-4 !p-8 border rounded !h-auto !text-center !border-cool-grey-400 dark:!border-dark-grey-500"
          >
            <Text variant="base" weight="strong">
              Existing GitHub connection
            </Text>
            <Text
              variant="body"
              className="text-balance text-center leading-relaxed"
            >
              Add an existing GitHub connection to your Nuon org by manually
              entering the GitHub{' '}
              <Badge className="!inline-block" variant="code" size="md">
                github_install_id
              </Badge>
            </Text>
          </Button>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          <Button
            variant="ghost"
            onClick={() => setIsManualMode(false)}
            className="!flex items-center gap-1.5 cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
          >
            <Icon variant="CaretLeft" />
            Back
          </Button>

          {error && (
            <Banner theme="error">
              {error?.error || 'Unable to create VCS connection.'}
            </Banner>
          )}

          <form
            ref={formRef}
            onSubmit={handleSubmit}
            className="flex flex-col gap-4"
          >
            <label className="flex flex-col gap-2">
              <Text variant="body" weight="strong">
                GitHub install ID
              </Text>
              <Input
                name="github_install_id"
                placeholder="github_installation_id"
                required
                type="text"
              />
            </label>
          </form>
        </div>
      )}
    </Modal>
  )
}

interface IConnectGithubButton extends IButtonAsButton {}

export const ConnectGithubButton = ({
  children = 'Add',
  ...props
}: IConnectGithubButton) => {
  const { addModal } = useSurfaces()
  const modal = <ConnectGithubModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      className="flex items-center gap-2 w-fit !border-transparent !p-2 !pl-1"
      size="sm"
      {...props}
    >
      <Icon variant="Plus" />
      {children}
    </Button>
  )
}
