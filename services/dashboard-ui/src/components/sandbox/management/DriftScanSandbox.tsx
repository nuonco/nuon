'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0'
import { reprovisionSandbox } from '@/actions/installs/reprovision-sandbox'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/segment-analytics'

export const DriftScanSandboxButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <DriftScanSandboxModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Scan" />}
      Drift scan sandbox
      {props?.isMenuButton ? <Icon variant="Scan" /> : null}
    </Button>
  )
}

export const DriftScanSandboxModal = ({
  ...props
}: IModal) => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const {
    data,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({ action: reprovisionSandbox })

  useServerActionToast({
    data,
    error,
    errorContent: <Text>Failed to initiate sandbox drift scan. Please try again.</Text>,
    errorHeading: `Sandbox drift scan failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: <Text>Sandbox drift scan workflow has been started successfully.</Text>,
    successHeading: `Drift scan initiated`,
  })

  const handleDriftScan = () => {
    execute({
      body: { plan_only: true },
      installId: install.id,
      orgId: org.id,
    })
  }

  const handleClose = () => {
    removeModal(props.modalId)
  }

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'install_sandbox_drift_scan',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: error?.error },
      })
    }

    if (data) {
      trackEvent({
        event: 'install_sandbox_drift_scan',
        user,
        status: 'ok',
        props: { orgId: org.id, installId: install.id },
      })

      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
    }
  }, [data, error, headers])

  return (
    <Modal
      heading="Drift scan sandbox?"
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Scanning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Scan" />
            Drift scan sandbox
          </span>
        ),
        disabled: isLoading,
        onClick: handleDriftScan,
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to drift scan sandbox.'}
          </Banner>
        ) : null}

        <Text variant="body" className="leading-relaxed">
          Are you sure you want to drift scan this sandbox?
        </Text>
      </div>
    </Modal>
  )
}