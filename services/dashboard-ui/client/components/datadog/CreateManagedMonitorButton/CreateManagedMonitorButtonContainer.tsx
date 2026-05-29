import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  createDatadogManagedMonitor,
  getDatadogConnections,
  getDatadogManagedMonitors,
} from '@/lib'
import type {
  TAPIError,
  TDatadogManagedMonitorPreset,
  TDatadogManagedMonitorTargetType,
} from '@/types'
import {
  CreateManagedMonitorModal,
  type CreateManagedMonitorInput,
} from './CreateManagedMonitorButton'

// CreateManagedMonitorButton is the "Alert in Datadog" surface embedded
// on resource detail pages (install, component, workflow — once action
// is wired). Two behaviors worth noting:
//
//  1. If the org has datadog disabled, render nothing. Don't show a
//     ghost button that does nothing on click.
//  2. If a managed monitor already exists for this (target_type,
//     target_id, preset='failure') tuple, swap the CTA copy from
//     "Alert in Datadog" → "Open existing alert" and link to DD via the
//     monitor's known shape. The dashboard never silently creates a
//     duplicate (the backend would idempotently no-op anyway, but the
//     UI confirms the existing row).
type Props = {
  targetType: TDatadogManagedMonitorTargetType
  targetId: string
  // installId qualifies action-target monitors so each install's
  // alert is scoped to that install's invocations. Required for
  // target_type="action" (the backend rejects org-wide action alerts
  // in v1); optional / ignored for the other target types whose
  // targetId is already install-scoped.
  installId?: string
  displayName?: string
  defaultPreset?: TDatadogManagedMonitorPreset
} & Omit<IButtonAsButton, 'children'>

const CreateManagedMonitorModalContainer = (
  props: {
    targetType: TDatadogManagedMonitorTargetType
    targetId: string
    installId?: string
    displayName?: string
    defaultPreset?: TDatadogManagedMonitorPreset
  } & Record<string, any>
) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const connectionsQuery = useQuery({
    queryKey: ['datadog-connections', org.id],
    queryFn: () => getDatadogConnections({ orgId: org.id }),
  })

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: CreateManagedMonitorInput) =>
      createDatadogManagedMonitor({
        orgId: org.id,
        body: {
          connection_id: input.connectionId,
          target_type: input.targetType,
          target_id: input.targetId,
          install_id: props.installId,
          preset: input.preset,
          display_name: input.displayName,
          notify_handles: input.notifyHandles,
        },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['datadog-managed-monitors', org.id],
      })
      addToast(
        <Toast heading="Datadog monitor created" theme="success">
          <Text>The monitor will fire on the next matching event.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to create monitor" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateManagedMonitorModal
      connections={connectionsQuery.data ?? []}
      targetType={props.targetType}
      targetId={props.targetId}
      displayName={props.displayName}
      defaultPreset={props.defaultPreset}
      isPending={isPending}
      error={error}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const CreateManagedMonitorButton = ({
  targetType,
  targetId,
  installId,
  displayName,
  defaultPreset,
  ...props
}: Props) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const hasDatadog = !!org?.features?.['datadog']

  // Existing-monitor lookup runs only when the org has DD enabled so we
  // don't fan out 404s for orgs that never use the integration.
  const monitorsQuery = useQuery({
    queryKey: ['datadog-managed-monitors', org?.id, targetId],
    queryFn: () =>
      getDatadogManagedMonitors({
        orgId: org!.id,
        targetId,
      }),
    enabled: !!org?.id && hasDatadog,
  })

  if (!hasDatadog) return null

  // Idempotency match must include install_id — the same action_workflow_id
  // can have one managed monitor per install. Compare with `|| ''` so a
  // missing install_id from either side normalizes to the same empty
  // string the backend uses for the non-action target types.
  const existing = monitorsQuery.data?.find(
    (m) =>
      m.target_type === targetType &&
      m.target_id === targetId &&
      (m.install_id || '') === (installId || '') &&
      (defaultPreset ? m.preset === defaultPreset : m.preset === 'failure')
  )

  if (existing) {
    // "Open existing alert" surfaces idempotency to the user. Clicking
    // wouldn't create a duplicate anyway, but a no-op create flow reads
    // worse than honesty about the row already existing.
    return (
      <Button
        variant="ghost"
        onClick={() =>
          addModal(
            <CreateManagedMonitorModalContainer
              targetType={targetType}
              targetId={targetId}
              installId={installId}
              displayName={displayName}
              defaultPreset={defaultPreset}
            />
          )
        }
        {...props}
      >
        <Icon variant="BellIcon" size={14} />
        Alert in Datadog (active)
      </Button>
    )
  }

  return (
    <Button
      variant="secondary"
      onClick={() =>
        addModal(
          <CreateManagedMonitorModalContainer
            targetType={targetType}
            targetId={targetId}
            installId={installId}
            displayName={displayName}
            defaultPreset={defaultPreset}
          />
        )
      }
      {...props}
    >
      <Icon variant="BellIcon" size={14} />
      Alert in Datadog
    </Button>
  )
}
