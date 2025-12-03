'use client'

import { usePathname, useRouter } from 'next/navigation'
import { syncSecrets } from '@/actions/installs/sync-secrets'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'

interface ISyncSecrets {}

export const SyncSecretsModal = ({ ...props }: ISyncSecrets & IModal) => {
  const path = usePathname()
  const router = useRouter()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data, error, headers, isLoading, execute } = useServerAction({
    action: syncSecrets,
  })

  useServerActionToast({
    data,
    error,
    errorContent: <Text>Unable to sync secrets for {install.name}.</Text>,
    errorHeading: `Secret sync failed`,
    onSuccess: () => {
      const workflowId = headers?.['x-nuon-install-workflow-id']
      const base = `/${org.id}/installs/${install.id}/workflows`
      const workflowPath = workflowId ? `${base}/${workflowId}` : base
      router.push(workflowPath)
      removeModal(props.modalId)
    },
    successContent: (
      <Text>Secrets for {install.name} are being synchronized.</Text>
    ),
    successHeading: `Secret sync started`,
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="Key" size="24" />
          Sync secrets?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Syncing secrets
          </span>
        ) : (
          'Sync secrets'
        ),
        onClick: () => {
          execute({
            orgId: org.id,
            path,
            installId: install.id,
            body: { plan_only: false }, // Empty body based on TSyncSecretsBody type
          })
        },
        disabled: isLoading,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="strong">
          Are you sure you want to sync secrets for {install.name}?
        </Text>
        <Text variant="base">
          This will synchronize all secrets from your app configuration to the
          install environment.
        </Text>
      </div>
    </Modal>
  )
}

export const SyncSecretsButton = ({
  ...props
}: ISyncSecrets & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <SyncSecretsModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Sync secrets
      <Icon variant="Key" />
    </Button>
  )
}
