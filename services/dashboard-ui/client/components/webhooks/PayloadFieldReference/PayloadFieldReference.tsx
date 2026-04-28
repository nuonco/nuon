import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'

type FieldRow = {
  field: string
  type: string
  presence: string
  description: React.ReactNode
}

const FIELD_COLUMNS = [
  {
    key: 'field' as const,
    header: 'Field',
    render: (value: FieldRow['field']) => (
      <Text variant="subtext" family="mono">{value}</Text>
    ),
  },
  {
    key: 'type' as const,
    header: 'Type',
    render: (value: FieldRow['type']) => (
      <Text variant="subtext" family="mono">{value}</Text>
    ),
  },
  {
    key: 'presence' as const,
    header: 'Presence',
    render: (value: FieldRow['presence']) => (
      <Text variant="subtext">{value}</Text>
    ),
  },
  {
    key: 'description' as const,
    header: 'Description',
    render: (_value: FieldRow['description'], item: FieldRow) => (
      <Text variant="subtext">{item.description}</Text>
    ),
  },
]

const ENVELOPE_FIELDS: FieldRow[] = [
  {
    field: 'specversion',
    type: 'string',
    presence: 'always',
    description: 'CloudEvents spec version. Currently always "1.0".',
  },
  {
    field: 'id',
    type: 'string (uuid)',
    presence: 'always',
    description:
      'Unique identifier for this delivery. A new id is generated for each event, even within the same operation.',
  },
  {
    field: 'type',
    type: 'string',
    presence: 'always',
    description: (
      <>
        Event type, always{' '}
        <span className="font-mono">com.nuon.operation.lifecycle.v1</span>.
      </>
    ),
  },
  {
    field: 'source',
    type: 'string (uri-reference)',
    presence: 'always',
    description: (
      <>
        Producer of the event, always{' '}
        <span className="font-mono">//nuon.co/ctl-api</span>.
      </>
    ),
  },
  {
    field: 'time',
    type: 'string (RFC 3339)',
    presence: 'always',
    description:
      'UTC timestamp at which the event was emitted by ctl-api.',
  },
  {
    field: 'subject',
    type: 'string',
    presence: 'always',
    description: (
      <>
        Stable, slash-joined identifier composed of org id, operation, stage,
        and event name (e.g.{' '}
        <span className="font-mono">
          org_…/component-deploy/plan/plan.finished
        </span>
        ).
      </>
    ),
  },
  {
    field: 'datacontenttype',
    type: 'string',
    presence: 'always',
    description: (
      <>
        Always <span className="font-mono">application/json</span>. Note the
        outer HTTP Content-Type is{' '}
        <span className="font-mono">application/cloudevents+json</span>.
      </>
    ),
  },
  {
    field: 'nuonorgid',
    type: 'string',
    presence: 'always',
    description: 'Org id this operation belongs to. Nuon CloudEvents extension.',
  },
  {
    field: 'nuonoperation',
    type: 'string',
    presence: 'always',
    description: (
      <>
        User-facing operation (e.g.{' '}
        <span className="font-mono">component-deploy</span>,{' '}
        <span className="font-mono">install-created</span>). Mirrors{' '}
        <span className="font-mono">data.operation</span>.
      </>
    ),
  },
  {
    field: 'nuonstage',
    type: 'string',
    presence: 'multi-phase only',
    description: (
      <>
        <span className="font-mono">plan</span> or{' '}
        <span className="font-mono">apply</span> for multi-phase operations.
        Omitted for single-phase operations.
      </>
    ),
  },
  {
    field: 'nuonstatus',
    type: 'string',
    presence: 'always',
    description: (
      <>
        One of <span className="font-mono">started</span>,{' '}
        <span className="font-mono">succeeded</span>,{' '}
        <span className="font-mono">failed</span>, or{' '}
        <span className="font-mono">canceled</span>. Mirrors{' '}
        <span className="font-mono">data.status</span>.
      </>
    ),
  },
  {
    field: 'data',
    type: 'object',
    presence: 'always',
    description: 'Nuon-specific payload — see fields below.',
  },
]

const DATA_FIELDS: FieldRow[] = [
  {
    field: 'data.event',
    type: 'string',
    presence: 'always',
    description: (
      <>
        Event name. One of <span className="font-mono">plan.started</span>,{' '}
        <span className="font-mono">plan.finished</span>,{' '}
        <span className="font-mono">apply.started</span>,{' '}
        <span className="font-mono">apply.finished</span>,{' '}
        <span className="font-mono">operation.started</span>, or{' '}
        <span className="font-mono">operation.finished</span>.
      </>
    ),
  },
  {
    field: 'data.operation',
    type: 'string',
    presence: 'always',
    description: (
      <>
        Mirror of <span className="font-mono">nuonoperation</span> on the
        envelope.
      </>
    ),
  },
  {
    field: 'data.stage',
    type: 'string',
    presence: 'multi-phase only',
    description: (
      <>
        <span className="font-mono">plan</span> or{' '}
        <span className="font-mono">apply</span> for multi-phase operations.
        Omitted for single-phase operations such as{' '}
        <span className="font-mono">install-created</span>.
      </>
    ),
  },
  {
    field: 'data.status',
    type: 'string',
    presence: 'always',
    description: (
      <>
        One of <span className="font-mono">started</span>,{' '}
        <span className="font-mono">succeeded</span>,{' '}
        <span className="font-mono">failed</span>, or{' '}
        <span className="font-mono">canceled</span>.
      </>
    ),
  },
  {
    field: 'data.failure_reason',
    type: 'string',
    presence: 'on failed only',
    description: (
      <>
        Set when the operation failed. Currently only{' '}
        <span className="font-mono">validation_failed</span> is emitted, when a
        stage's pre-execution validation rejects the operation.
      </>
    ),
  },
  {
    field: 'data.error',
    type: 'string',
    presence: 'on failed only',
    description:
      'Human-readable error message. Omitted when the operation succeeded or was canceled.',
  },
  {
    field: 'data.duration_ms',
    type: 'integer',
    presence: 'on *.finished events',
    description: 'How long the operation took to run, in milliseconds.',
  },
  {
    field: 'data.metadata',
    type: 'object',
    presence: 'optional',
    description:
      'Additional operation-specific metadata. Shape varies by operation; treat unknown keys as opaque.',
  },
]

const CONTEXT_FIELDS: FieldRow[] = [
  {
    field: 'data.context.org_id',
    type: 'string',
    presence: 'always',
    description: 'Org id this operation belongs to.',
  },
  {
    field: 'data.context.install_id',
    type: 'string | omitted',
    presence: 'when relevant',
    description:
      'Install id when the operation targets an install (install create/update/restart, sandbox provisioning, etc.).',
  },
  {
    field: 'data.context.component_id',
    type: 'string | omitted',
    presence: 'when relevant',
    description:
      'Component id when the operation targets a component (component deploy/teardown).',
  },
  {
    field: 'data.context.sandbox_id',
    type: 'string | omitted',
    presence: 'when relevant',
    description: 'Sandbox id for sandbox provisioning operations.',
  },
]

const GRID_TEMPLATE = 'minmax(150px, 2fr) minmax(120px, 1.5fr) minmax(100px, 1fr) minmax(200px, 3fr)'

export const PayloadFieldReference = () => (
  <div className="flex flex-col gap-6">
    <div className="flex flex-col gap-2">
      <Text variant="base" weight="strong">
        Field reference
      </Text>
      <Text variant="body" theme="neutral">
        Every delivery uses the same shape. Each operation produces a{' '}
        <span className="font-mono">*.started</span> event when it begins and a{' '}
        <span className="font-mono">*.finished</span> event when it completes.
        Multi-phase operations (component deploy/teardown, sandbox
        provision/reprovision/deprovision) emit{' '}
        <span className="font-mono">plan.*</span> and{' '}
        <span className="font-mono">apply.*</span> events; single-phase
        operations emit <span className="font-mono">operation.*</span> events.
      </Text>
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        CloudEvents envelope
      </Text>
      <PropertyGrid
        className="rounded-md border p-4"
        columns={FIELD_COLUMNS}
        values={ENVELOPE_FIELDS}
        gridTemplate={GRID_TEMPLATE}
      />
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        data
      </Text>
      <Text variant="subtext" theme="neutral">
        Describes the operation, its stage, and the outcome.
      </Text>
      <PropertyGrid
        className="rounded-md border p-4"
        columns={FIELD_COLUMNS}
        values={DATA_FIELDS}
        gridTemplate={GRID_TEMPLATE}
      />
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        data.context
      </Text>
      <Text variant="subtext" theme="neutral">
        Resource ids the operation applies to. Only the ids relevant to the
        operation are included; the rest are omitted.
      </Text>
      <PropertyGrid
        className="rounded-md border p-4"
        columns={FIELD_COLUMNS}
        values={CONTEXT_FIELDS}
        gridTemplate={GRID_TEMPLATE}
      />
    </div>
  </div>
)
