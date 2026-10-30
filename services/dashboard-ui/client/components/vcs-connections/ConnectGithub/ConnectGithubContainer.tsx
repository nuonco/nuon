import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { ConnectGithubModal } from './ConnectGithub'

export const ConnectGithubModalContainer = ({
  onSubmit: _onSubmit,
  ...props
}: IModal) => {
  const { githubAppName } = useConfig()
  const { org } = useOrg()

  return (
    <ConnectGithubModal githubAppName={githubAppName} orgId={org.id} {...props} />
  )
}

export const ConnectGithubButton = ({
  children = 'Add',
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ConnectGithubModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="PlusIcon" />}
      {children}
      {props?.isMenuButton ? <Icon variant="PlusIcon" /> : null}
    </Button>
  )
}
