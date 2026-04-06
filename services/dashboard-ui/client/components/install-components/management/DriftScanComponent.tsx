import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deployComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'
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
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const [buildId, setBuildId] = useState<string>()

  const { mutate: execute, isPending: isLoading, error } = useMutation({
    mutationFn: (params: { body: Parameters<typeof deployComponent>[0]['body'] }) =>
      deployComponent({
        body: params.body,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
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
      addToast(
        <Toast heading={`${component.name} drift scan started`} theme="success">
          <Text>Drift scan for {component.name} was started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (err: any) => {
      trackEvent({
        event: 'component_drift_scan',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
          buildId,
          err: err?.error,
        },
      })
      addToast(
        <Toast heading="Component drift scan failed" theme="error">
          <Text>Unable to drift scan {component.name} component.</Text>
        </Toast>
      )
    },
  })

  const handleClose = () => {
    setBuildId(undefined)
    removeModal(props.modalId)
  }

  const handleBuildSelect = (selectedBuildId: string) => {
    setBuildId(selectedBuildId)
  }

  const isDriftScanDisabled = !buildId || isLoading

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
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
      size="half"
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      onClose={handleClose}
      primaryActionTrigger={{
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
          })
        },
        variant: 'primary' as const,
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to drift scan component'}
          </Banner>
        ) : null}

        <BuildSelect
          componentId={component.id}
          componentType={component.type}
          selectedBuildId={buildId}
          currentBuildId={currentBuildId}
          onSelectBuild={handleBuildSelect}
          onClose={handleClose}
        />
      </div>
    </Modal>
  )
}
