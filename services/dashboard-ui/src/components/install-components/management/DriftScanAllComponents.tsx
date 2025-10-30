'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { deployComponents } from '@/actions/installs/deploy-components'
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

export const DriftScanAllComponentsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DriftScanAllComponentsModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Drift scan components <Icon variant="ScanIcon" />
    </Button>
  )
}

export const DriftScanAllComponentsModal = ({ ...props }: IModal) => {
  const path = usePathname()
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const [isKickedOff, setIsKickedOff] = useState(false)

  const {
    data: deploysOk,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({ action: deployComponents })

  useServerActionToast({
    data: deploysOk,
    error,
    errorContent: (
      <Text>Unabled to deploy all components to {install.name}.</Text>
    ),
    errorHeading: `Drift scan failed`,
    onSuccess: () => {
      const workflowId = headers?.['x-nuon-install-workflow-id']
      const base = `/${org.id}/installs/${install.id}/workflows`
      const workflowPath = workflowId ? `${base}/${workflowId}` : base
      router.push(workflowPath)
      removeModal(props.modalId)
    },
    successContent: <Text>Components drift scan workflow was created.</Text>,
    successHeading: `${install.name} component drift scan started`,
  })

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'components_drift_scan',
        status: 'error',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
          err: error?.error,
        },
      })
    }

    if (deploysOk) {
      trackEvent({
        event: 'components_drift_scan',
        status: 'ok',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
        },
      })

      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers?.['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
    }
  }, [deploysOk, error, headers])

  return (
    <Modal
      heading={
        <Text
          className="!inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="ScanIcon" size="24" />
          Drift scan all {install.name} components?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting drift scan
          </span>
        ) : (
          'Drift scan components'
        ),
        disabled: isKickedOff,
        onClick: () => {
          execute({
            orgId: org.id,
            path,
            installId: install.id,
            body: { plan_only: true },
          })
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to deploy components'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to drift scan components?
        </Text>
        <Text variant="base">
          This aciton will preform a drift scan against the latest build of each component and the component deployments on your
          install.
        </Text>
      </div>
    </Modal>
  )
}
