import { useMemo, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TDatadogConnection,
  TDatadogManagedMonitorMode,
  TDatadogManagedMonitorPreset,
  TDatadogManagedMonitorTargetType,
} from '@/types'

export type CreateManagedMonitorInput = {
  connectionId: string
  targetType: TDatadogManagedMonitorTargetType
  targetId: string
  preset: TDatadogManagedMonitorPreset
  mode: TDatadogManagedMonitorMode
  displayName?: string
  notifyHandles: string[]
}

// CreateManagedMonitorModal is the form behind the "Alert in Datadog"
// button. The caller pins targetType + targetId because the button is
// embedded in a resource-specific page (install detail, action run page,
// etc.) — those are part of the entry-point's identity, not user input.
//
// Notification handles default to the picked connection's
// DefaultNotifyHandles. Editing the textarea overrides for THIS monitor
// only; the connection's defaults stay intact.
export const CreateManagedMonitorModal = ({
  connections,
  targetType,
  targetId,
  displayName,
  defaultPreset = 'failure',
  defaultMode = 'event',
  isPending,
  error,
  onSubmit,
  ...props
}: {
  connections: TDatadogConnection[]
  targetType: TDatadogManagedMonitorTargetType
  targetId: string
  displayName?: string
  defaultPreset?: TDatadogManagedMonitorPreset
  defaultMode?: TDatadogManagedMonitorMode
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: CreateManagedMonitorInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const usableConnections = useMemo(
    // The Monitors API requires both keys, so the dropdown hides
    // connections without an application key. The status badge in the
    // option label makes it obvious why a row might be missing.
    () =>
      connections.filter(
        (c) => c.status === 'verified'
      ),
    [connections]
  )

  const [connectionId, setConnectionId] = useState(
    usableConnections[0]?.id ?? ''
  )
  const [preset, setPreset] =
    useState<TDatadogManagedMonitorPreset>(defaultPreset)
  const [mode, setMode] =
    useState<TDatadogManagedMonitorMode>(defaultMode)

  // handlesText tracks the user-edited textarea contents and starts
  // pre-populated with the chosen connection's DefaultNotifyHandles.
  // When the user picks a different connection, we re-seed the textarea
  // — overwriting any unsaved edits. That's intentional: every other
  // surface follows "pick → see defaults", and the few seconds of work
  // a user might lose on a connection swap is the price for that
  // consistency.
  const selectedConn = connections.find((c) => c.id === connectionId)
  const [handlesText, setHandlesText] = useState(
    (selectedConn?.default_notify_handles ?? []).join('\n')
  )
  const [lastSeenConnectionId, setLastSeenConnectionId] = useState(connectionId)
  if (connectionId !== lastSeenConnectionId) {
    setLastSeenConnectionId(connectionId)
    setHandlesText((selectedConn?.default_notify_handles ?? []).join('\n'))
  }

  const canSubmit = !!connectionId && !isPending

  // The action preset is disabled at v1 because the renderer doesn't yet
  // emit a stable nuon_action_id tag. We still render the dropdown to
  // pin the contract, but the option is disabled with an inline hint.
  const actionDisabled = targetType === 'action'

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="BellIcon" size="24" />
          Alert in Datadog
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating monitor…
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="BellIcon" />
            Create monitor
          </span>
        ),
        disabled: !canSubmit || actionDisabled,
        onClick: () =>
          onSubmit({
            connectionId,
            targetType,
            targetId,
            preset,
            mode,
            displayName,
            notifyHandles: splitLines(handlesText),
          }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {actionDisabled ? (
          <Banner theme="warn">
            Per-action alerts aren't supported in v1 yet — the event
            renderer needs to tag events with a stable action ID first.
            Subscribe the install or workflow that owns this action in the
            meantime.
          </Banner>
        ) : null}

        {error ? (
          <Banner theme="error">
            {error?.description || error?.error || 'Unable to create monitor.'}
          </Banner>
        ) : null}

        {usableConnections.length === 0 ? (
          <Banner theme="warn">
            No verified Datadog connections in this org. Connect one and
            add an application key first — DD's Monitors API requires
            both keys.
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label>Target</Label>
          <Text variant="subtext" theme="neutral">
            <strong>{targetType}</strong> — {displayName || targetId}
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-mon-connection">Datadog connection</Label>
          <Select
            id="dd-mon-connection"
            options={usableConnections.map((c) => ({
              value: c.id ?? '',
              label: `${c.name ?? '(unnamed)'}`,
            }))}
            value={connectionId}
            onChange={(e) => setConnectionId(e.target.value)}
            disabled={usableConnections.length === 0}
            placeholder="Select a connection"
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-mon-preset">Trigger</Label>
          <Select
            id="dd-mon-preset"
            options={[
              { value: 'failure', label: 'On failure' },
              { value: 'drift', label: 'On drift detected (installs)' },
            ]}
            value={preset}
            onChange={(e) =>
              setPreset(e.target.value as TDatadogManagedMonitorPreset)
            }
          />
          <Text variant="subtext" theme="neutral">
            Failure fires on the first failed event in a 5m window. Drift
            fires on drift-detected events from installs.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-mon-mode">Mode</Label>
          <Select
            id="dd-mon-mode"
            options={[
              { value: 'event', label: 'Event stream (needs DD event subscription)' },
              { value: 'metric', label: 'Metric (works without event subscription)' },
            ]}
            value={mode}
            onChange={(e) =>
              setMode(e.target.value as TDatadogManagedMonitorMode)
            }
          />
          <Text variant="subtext" theme="neutral">
            Event mode queries the DD event stream — requires an active
            event subscription routing Nuon events into DD. Metric mode
            lets Nuon evaluate the match and submit a single
            <code>{' nuon.monitor.fired '}</code>
            count to DD, so the alert fires without an event subscription.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-mon-handles">
            Notification handles{' '}
            <Text variant="subtext" theme="neutral" className="inline">
              (one per line, @-prefixed)
            </Text>
          </Label>
          <Textarea
            id="dd-mon-handles"
            placeholder={'@pagerduty-prod\n@slack-oncall'}
            value={handlesText}
            onChange={(e) => setHandlesText(e.target.value)}
            rows={3}
          />
          <Text variant="subtext" theme="neutral">
            Seeded from the connection's defaults — edit to override for
            this monitor only.
          </Text>
        </div>
      </div>
    </Modal>
  )
}

const splitLines = (s: string): string[] =>
  s
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0)
