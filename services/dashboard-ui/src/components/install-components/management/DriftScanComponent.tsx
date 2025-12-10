'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { deployComponent } from '@/actions/installs/deploy-component'
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
import type { TComponent } from '@/types'
import { trackEvent } from '@/lib/segment-analytics'
import { BuildSelect } from './BuildSelect'

export const DriftScanComponentButton = ({
  component,
  currentBuildId,
  ...props
}: IButtonAsButton & {
  component: TComponent
  currentBuildId?: string
}) => {
  const { addModal } = useSurfaces()
  const modal = <DriftScanComponentModal component={component} currentBuildId={currentBuildId} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Scan" />}
      Drift scan component
      {props?.isMenuButton ? <Icon variant="Scan" /> : null}
    </Button>
  )
}

export const DriftScanComponentModal = ({
  component,
  currentBuildId,
  ...props
}: IModal & {
  component: TComponent
  currentBuildId?: string
}) => {
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const [buildId, setBuildId] = useState<string>()

  const {
    data: deploy,
    error,
    execute,
    isLoading,
    headers,
  } = useServerAction({ action: deployComponent })

  useServerActionToast({
    data: deploy,
    error,
    errorContent: <Text>Unable to drift scan {component.name} component.</Text>,
    errorHeading: `Component drift scan failed`,
    onSuccess: () => {
      removeModal(props.modalId)
      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    successContent: <Text>Drift scan for {component.name} was started.</Text>,
    successHeading: `${component.name} drift scan started`,
  })

  const handleClose = () => {
    setBuildId(undefined)
    removeModal(props.modalId)
  }

  const handleBuildSelect = (selectedBuildId: string) => {
    setBuildId(selectedBuildId)
  }

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'component_drift_scan',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
          buildId,
          err: error?.error,
        },
      })
    }

    if (deploy) {
      trackEvent({
        event: 'component_drift_scan',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
          buildId,
        },
      })
    }
  }, [deploy, error, org.id, install.id, component.id, buildId, user])

  const isDriftScanDisabled = !buildId || isLoading

  // Always show the drift scan button, disabled when no build selected
  const modalProps = {
    primaryActionTrigger: {
      children: isLoading ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" />
          Scanning build
        </span>
      ) : (
        <span className="flex items-center gap-2">
          <Icon variant="MagnifyingGlass" />
          Drift scan build
        </span>
      ),
      disabled: isDriftScanDisabled,
      onClick: () => {
        execute({
          body: {
            build_id: buildId!,
            deploy_dependents: false,
            plan_only: true,
          },
          installId: install.id,
          orgId: org.id,
        })
      },
      variant: 'primary' as const,
    },
  }

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            className="inline-flex gap-4 items-center"
            variant="h3"
            weight="strong"
            theme="info"
          >
            <Icon variant="Scan" size="24" />
            Drift scan {component.name} component
          </Text>
          <Text
            variant="body"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            Select a build to scan for drift
          </Text>
        </div>
      }
      size="3/4"
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      onClose={handleClose}
      {...modalProps}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to drift scan component'}
          </Banner>
        ) : null}

        <BuildSelect
          componentId={component.id}
          selectedBuildId={buildId}
          currentBuildId={currentBuildId}
          onSelectBuild={handleBuildSelect}
          onClose={handleClose}
        />
      </div>
    </Modal>
  )
}
