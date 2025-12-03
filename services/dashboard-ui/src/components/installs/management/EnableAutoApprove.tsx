'use client'

import { useState } from 'react'
import { usePathname } from 'next/navigation'
import { createInstallConfig } from '@/actions/installs/create-install-config'
import { updateInstallConfig } from '@/actions/installs/update-install-config'
import { updateInstall } from '@/actions/installs/update-install'
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

interface IEnableAutoApprove {}

// Confirm Override Modal for installs managed by config
export const ConfirmOverrideModal = ({ onConfirm, ...props }: { onConfirm: () => void } & IModal) => {
  const { removeModal } = useSurfaces()
  const { install } = useInstall()

  const isInstallManagedByConfig = install?.metadata?.managed_by === 'nuon/cli/install-config'

  if (!isInstallManagedByConfig) {
    return null
  }

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="warn"
        >
          <Icon variant="Warning" size="24" />
          Override changes to this install?
        </Text>
      }
      primaryActionTrigger={{
        children: 'Confirm override',
        onClick: () => {
          onConfirm()
          removeModal(props.modalId)
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          <Text variant="base" weight="strong">
            You are about to update an Install managed by a Config file.
          </Text>
          <Text variant="body">
            If you proceed, the config file syncing will be disabled. Are you sure you want to continue?
          </Text>
        </div>
        <Banner theme="info">
          <Text variant="body">
            <strong>Tip:</strong> Use the management menu to enable Install Config syncing again.
          </Text>
        </Banner>
      </div>
    </Modal>
  )
}

export const EnableAutoApproveModal = ({ ...props }: IEnableAutoApprove & IModal) => {
  const path = usePathname()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()

  const hasInstallConfig = Boolean(install?.install_config)
  const isApproveAll = hasInstallConfig && install?.install_config?.approval_option === 'approve-all'

  // Create install config action (for installs without config)
  const { data: createData, error: createError, isLoading: createLoading, execute: executeCreate } = useServerAction({
    action: createInstallConfig,
  })

  // Update install config action (for installs with existing config)  
  const { data: updateData, error: updateError, isLoading: updateLoading, execute: executeUpdate } = useServerAction({
    action: updateInstallConfig,
  })

  // Update install metadata action (to change managed_by)
  const { execute: executeInstallUpdate } = useServerAction({
    action: updateInstall,
  })

  // Combined toast handling for both create and update actions
  const isLoading = createLoading || updateLoading
  const data = createData || updateData
  const error = createError || updateError

  useServerActionToast({
    data,
    error,
    errorContent: (
      <Text>Unable to {isApproveAll ? 'disable' : 'enable'} auto approve for {install.name}.</Text>
    ),
    errorHeading: `Auto approve ${isApproveAll ? 'disable' : 'enable'} failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: (
      <Text>Auto approve {isApproveAll ? 'disabled' : 'enabled'} for {install.name}.</Text>
    ),
    successHeading: `Auto approve ${isApproveAll ? 'disabled' : 'enabled'}`,
  })

  const handleToggleApproval = async () => {
    // If install is managed by CLI config, update metadata first
    if (install?.metadata?.managed_by === 'nuon/cli/install-config') {
      await executeInstallUpdate({
        orgId: org.id,
        path,
        installId: install.id,
        body: { metadata: { managed_by: 'nuon/dashboard' } },
      })
    }

    if (hasInstallConfig) {
      // Update existing config
      executeUpdate({
        orgId: org.id,
        path,
        installId: install.id,
        installConfigId: install.install_config.id,
        body: { approval_option: isApproveAll ? 'prompt' : 'approve-all' },
      })
    } else {
      // Create new config with auto approve enabled
      executeCreate({
        orgId: org.id,
        path,
        installId: install.id,
        body: { approval_option: 'approve-all' },
      })
    }
  }

  const buttonText = isApproveAll ? 'Disable auto approval' : 'Enable auto approval'
  const confirmText = isApproveAll 
    ? 'Are you sure you want to disable auto approve for changes to this install?'
    : 'Are you sure you want to enable auto approve for changes to this install?'

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant={isApproveAll ? "ToggleRight" : "ToggleLeft"} size="24" />
          {buttonText}?
        </Text>
      }
      primaryActionTrigger={{
        children: data ? (
          <span className="flex items-center gap-2">
            <Icon variant="CheckCircle" /> Settings updated
          </span>
        ) : isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Updating settings...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant={isApproveAll ? "ToggleRight" : "ToggleLeft"} />
            {buttonText}
          </span>
        ),
        onClick: handleToggleApproval,
        disabled: isLoading || Boolean(data),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || `Unable to ${isApproveAll ? 'disable' : 'enable'} auto approval`}
          </Banner>
        ) : null}

        <Text variant="body">
          {confirmText}
        </Text>

        {!isApproveAll && (
          <Banner theme="warn">
            <Text variant="body">
              <strong>Warning:</strong> When auto approve is enabled, all changes to this install will be automatically approved and applied without manual review.
            </Text>
          </Banner>
        )}
      </div>
    </Modal>
  )
}

export const EnableAutoApproveButton = ({
  ...props
}: IEnableAutoApprove & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const { install } = useInstall()

  const hasInstallConfig = Boolean(install?.install_config)
  const isApproveAll = hasInstallConfig && install?.install_config?.approval_option === 'approve-all'
  const isInstallManagedByConfig = install?.metadata?.managed_by === 'nuon/cli/install-config'

  const buttonText = isApproveAll ? 'Disable auto approval' : 'Enable auto approval'
  const buttonIcon = isApproveAll ? 'ToggleRight' : 'ToggleLeft'

  const handleClick = () => {
    if (isInstallManagedByConfig) {
      // Show override warning first
      const overrideModal = (
        <ConfirmOverrideModal
          onConfirm={() => {
            // After confirmation, show the main modal
            const mainModal = <EnableAutoApproveModal />
            addModal(mainModal)
          }}
        />
      )
      addModal(overrideModal)
    } else {
      // Show main modal directly
      const modal = <EnableAutoApproveModal />
      addModal(modal)
    }
  }

  return (
    <Button
      onClick={handleClick}
      {...props}
    >
      {buttonText}
      <Icon variant={buttonIcon} />
    </Button>
  )
}