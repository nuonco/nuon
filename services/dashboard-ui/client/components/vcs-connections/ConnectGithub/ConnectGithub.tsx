import React from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

export const githubConnectStartUrl = ({
  orgId,
  onboarding = false,
}: {
  orgId: string
  onboarding?: boolean
}) => {
  const params = new URLSearchParams({ org_id: orgId })
  if (onboarding) params.set('onboarding', '1')
  return `/api/connect-github?${params}`
}

interface IConnectGithubModal extends Omit<IModal, 'onSubmit'> {
  githubAppName: string
  orgId: string
}

// Connections are only ever created through the GitHub install redirect, which
// lets ctl-api confirm the user owns the installation. There is no manual
// install-ID entry: an id alone can be spoofed. To attach an installation that
// already exists, pick that account on GitHub's install page and it redirects
// back the same way.
export const ConnectGithubModal = ({
  githubAppName,
  orgId,
  ...props
}: IConnectGithubModal) => {
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
      {...props}
    >
      <div className="flex flex-col gap-6">
        <Button
          href={githubConnectStartUrl({ orgId })}
          isAnchorTag
          variant="ghost"
          className="flex flex-col items-center justify-center gap-4 !p-8 rounded !h-auto !text-center !border-cool-grey-400 dark:!border-dark-grey-500"
        >
          <Text variant="base" weight="strong">
            Connect a GitHub account
          </Text>
          <Text
            variant="body"
            className="!inline-block text-balance !text-center leading-relaxed"
          >
            Install the{' '}
            <Badge className="!inline-block" variant="code" size="md">
              {githubAppName}
            </Badge>{' '}
            GitHub app and allow access to the repositories of your choice. If
            the app is already installed, choose that account to connect it to
            this org.
          </Text>
        </Button>
      </div>
    </Modal>
  )
}
