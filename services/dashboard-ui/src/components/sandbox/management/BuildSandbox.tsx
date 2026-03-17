'use client'

import { usePathname } from 'next/navigation'
import { createSandboxBuild } from '@/actions/apps/create-sandbox-build'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'

export const BuildSandboxButton = (props: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <BuildSandboxModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Hammer" />}
      Build sandbox
      {props?.isMenuButton ? <Icon variant="Hammer" /> : null}
    </Button>
  )
}

export const BuildSandboxModal = (props: IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()

  const {
    data: build,
    error,
    execute,
    isLoading,
  } = useServerAction({ action: createSandboxBuild })

  useServerActionToast({
    data: build,
    error,
    errorContent: <Text>Unable to build sandbox.</Text>,
    errorHeading: 'Sandbox build failed',
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: <Text>Sandbox build was started.</Text>,
    successHeading: 'Sandbox build started',
  })

  return (
    <Modal
      heading={
        <Text
          className="!inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="Hammer" size="24" />
          Build sandbox?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Building sandbox
          </span>
        ) : (
          'Build sandbox'
        ),
        disabled: isLoading,
        onClick: () => {
          execute({
            appId: app.id,
            orgId: org.id,
            path,
          })
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to build sandbox'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to build the sandbox?
        </Text>
        <Text variant="base">
          This will start a standalone sandbox build using the latest sandbox
          configuration. The build process may take several minutes to complete.
        </Text>
      </div>
    </Modal>
  )
}
