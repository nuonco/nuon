import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deployComponent } from '@/lib'
import { trackEvent } from '@/lib/posthog-analytics'
import type { TComponent } from '@/types'
import { BuildSelect } from '../BuildSelect'
import { DriftScanComponentModal } from './DriftScanComponent'

export const DriftScanComponentModalContainer = ({
  component,
  currentBuildId,
  ...props
}: Omit<IModal, 'onSubmit'> & {
  component: TComponent
  currentBuildId?: string
}) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate: execute, isPending, error } = useMutation({
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
        },
      })
      addToast(
        <Toast heading={`${component.name} drift scan started`} theme="success">
          <Text>Drift scan for {component.name} was started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
      if (workflowId) {
        navigate(installLink({ installId: install.id, appId: install.app_id, suffix: `/workflows/${workflowId}` }))
      } else {
        navigate(installLink({ installId: install.id, appId: install.app_id, suffix: `/workflows` }))
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

  return (
    <DriftScanComponentModal
      component={component}
      currentBuildId={currentBuildId}
      isPending={isPending}
      error={error as any}
      onSubmit={({ buildId }) => {
        execute({
          body: {
            build_id: buildId,
            deploy_dependents: false,
            plan_only: true,
          },
        })
      }}
      onClose={() => removeModal(props.modalId)}
      buildSelect={({ selectedBuildId, onSelectBuild, onClose }) => (
        <BuildSelect
          componentId={component.id}
          componentType={component.type}
          selectedBuildId={selectedBuildId}
          currentBuildId={currentBuildId}
          onSelectBuild={onSelectBuild}
          onClose={onClose}
        />
      )}
      {...props}
    />
  )
}

export const DriftScanComponentButton = ({
  component,
  currentBuildId,
  ...props
}: IButtonAsButton & {
  component: TComponent
  currentBuildId?: string
}) => {
  const { addModal } = useSurfaces()
  const modal = <DriftScanComponentModalContainer component={component} currentBuildId={currentBuildId} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="ScanIcon" />}
      Drift scan component
      {props?.isMenuButton ? <Icon variant="ScanIcon" /> : null}
    </Button>
  )
}
