import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createVCSConnection } from '@/lib/ctl-api/vcs-connections'
import { ConnectGithubModal } from './ConnectGithub'

export const ConnectGithubModalContainer = ({
  onSubmit: _onSubmit,
  ...props
}: IModal) => {
  const { githubAppName } = useConfig()
  const { org, refresh: refreshOrg } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: createVCSConnection,
    onSuccess: () => {
      refreshOrg()
      addToast(
        <Toast theme="info" heading="GitHub connected">
          <Text>Your GitHub connection has been added.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err) => {
      addToast(
        <Toast theme="error" heading="GitHub connection failed">
          <Text variant="subtext">{err?.error}</Text>
        </Toast>
      )
    },
  })

  return (
    <ConnectGithubModal
      githubAppName={githubAppName}
      orgId={org.id}
      isPending={isPending}
      error={error}
      onSubmit={(githubInstallId) =>
        mutate({
          body: { github_install_id: githubInstallId },
          orgId: org.id,
        })
      }
      {...props}
    />
  )
}

interface IConnectGithubButton extends IButtonAsButton {
  isIconFirst?: boolean
}

export const ConnectGithubButton = ({
  children = 'Add',
  isIconFirst = false,
  ...props
}: IConnectGithubButton) => {
  const { addModal } = useSurfaces()
  const modal = <ConnectGithubModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {isIconFirst ? <Icon variant="PlusIcon" /> : null}
      {children}
      {!isIconFirst ? <Icon variant="PlusIcon" /> : null}
    </Button>
  )
}
