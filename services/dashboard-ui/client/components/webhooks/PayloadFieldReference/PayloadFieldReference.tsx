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
      'Unique identifier for this delivery. A new id is generated for each event, even within the same workflow.',
  },
  {
    field: 'type',
    type: 'string',
    presence: 'always',
    description: (
      <>
        Event type. One of{' '}
        <span className="font-mono">com.nuon.workflow.lifecycle.v1</span> or{' '}
        <span className="font-mono">com.nuon.workflow_step.lifecycle.v1</span>.
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
        Stable, slash-joined identifier composed of org id, kind, workflow id,
        optional step id, and transition (e.g.{' '}
        <span className="font-mono">
          org_…/workflow_step/inwYY…/iws_…/succeeded
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
    description: 'Org id this event belongs to. Nuon CloudEvents extension.',
  },
  {
    field: 'nuonkind',
    type: 'string',
    presence: 'always',
    description: (
      <>
        One of <span className="font-mono">workflow</span> or{' '}
        <span className="font-mono">workflow_step</span>. Mirrors{' '}
        <span className="font-mono">data.kind</span>.
      </>
    ),
  },
  {
    field: 'nuontransition',
    type: 'string',
    presence: 'always',
    description: (
      <>
        One of <span className="font-mono">started</span>,{' '}
        <span className="font-mono">succeeded</span>,{' '}
        <span className="font-mono">failed</span>, or{' '}
        <span className="font-mono">cancelled</span>. Mirrors{' '}
        <span className="font-mono">data.transition</span>.
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
    field: 'data.kind',
    type: 'string',
    presence: 'always',
    description: (
      <>
        One of <span className="font-mono">workflow</span> or{' '}
        <span className="font-mono">workflow_step</span>.
      </>
    ),
  },
  {
    field: 'data.transition',
    type: 'string',
    presence: 'always',
    description: (
      <>
        One of <span className="font-mono">started</span>,{' '}
        <span className="font-mono">succeeded</span>,{' '}
        <span className="font-mono">failed</span>, or{' '}
        <span className="font-mono">cancelled</span>.
      </>
    ),
  },
  {
    field: 'data.org_id',
    type: 'string',
    presence: 'always',
    description: 'Org id this event belongs to.',
  },
]

const WORKFLOW_FIELDS: FieldRow[] = [
  {
    field: 'data.workflow.id',
    type: 'string',
    presence: 'always',
    description: 'Workflow id. Stable across all events for the workflow.',
  },
  {
    field: 'data.workflow.type',
    type: 'string',
    presence: 'always',
    description: (
      <>
        Workflow kind (e.g. <span className="font-mono">provision</span>,{' '}
        <span className="font-mono">reprovision</span>,{' '}
        <span className="font-mono">manual_deploy</span>,{' '}
        <span className="font-mono">action_workflow_run</span>).
      </>
    ),
  },
  {
    field: 'data.workflow.owner_id',
    type: 'string',
    presence: 'always',
    description:
      'Id of the entity that owns the workflow (an install, app, or app branch).',
  },
  {
    field: 'data.workflow.owner_type',
    type: 'string',
    presence: 'always',
    description: (
      <>
        One of <span className="font-mono">installs</span>,{' '}
        <span className="font-mono">apps</span>, or{' '}
        <span className="font-mono">app_branches</span>.
      </>
    ),
  },
]

const STEP_FIELDS: FieldRow[] = [
  {
    field: 'data.step.id',
    type: 'string',
    presence: 'workflow_step events only',
    description: 'Workflow step id.',
  },
  {
    field: 'data.step.name',
    type: 'string',
    presence: 'workflow_step events only',
    description:
      'Human-readable step name, e.g. "deploy api (apply)".',
  },
  {
    field: 'data.step.idx',
    type: 'integer',
    presence: 'workflow_step events only',
    description: 'Step index within the workflow.',
  },
  {
    field: 'data.step.target_type',
    type: 'string',
    presence: 'workflow_step events only',
    description: (
      <>
        Resource the step manipulates (e.g.{' '}
        <span className="font-mono">install_deploys</span>,{' '}
        <span className="font-mono">install_sandbox_runs</span>,{' '}
        <span className="font-mono">install_action_workflow_runs</span>).
      </>
    ),
  },
  {
    field: 'data.step.target_id',
    type: 'string',
    presence: 'workflow_step events only',
    description: 'Id of the manipulated resource.',
  },
  {
    field: 'data.step.component_id',
    type: 'string | omitted',
    presence: 'when target is a deploy',
    description: (
      <>
        Component id, set when{' '}
        <span className="font-mono">target_type</span> is{' '}
        <span className="font-mono">install_deploys</span>.
      </>
    ),
  },
  {
    field: 'data.step.sandbox_id',
    type: 'string | omitted',
    presence: 'when target is a sandbox run',
    description: (
      <>
        Sandbox id, set when{' '}
        <span className="font-mono">target_type</span> is{' '}
        <span className="font-mono">install_sandbox_runs</span>.
      </>
    ),
  },
  {
    field: 'data.step.execution_type',
    type: 'string',
    presence: 'workflow_step events only',
    description: (
      <>
        One of <span className="font-mono">system</span>,{' '}
        <span className="font-mono">user</span>,{' '}
        <span className="font-mono">approval</span>,{' '}
        <span className="font-mono">skipped</span>, or{' '}
        <span className="font-mono">hidden</span>.
      </>
    ),
  },
]

const PARENT_FIELDS: FieldRow[] = [
  {
    field: 'data.parent.workflow_id',
    type: 'string',
    presence: 'when nested',
    description: 'Parent workflow id when this workflow was launched from another workflow\'s step.',
  },
  {
    field: 'data.parent.step_id',
    type: 'string',
    presence: 'when nested',
    description: 'Parent step id (the step in the parent workflow that triggered this workflow).',
  },
  {
    field: 'data.parent.kind',
    type: 'string',
    presence: 'when nested',
    description: (
      <>
        Always <span className="font-mono">workflow_step</span>.
      </>
    ),
  },
]

const OUTCOME_FIELDS: FieldRow[] = [
  {
    field: 'data.outcome.status',
    type: 'string',
    presence: 'on terminal events',
    description: (
      <>
        Mirrors <span className="font-mono">data.transition</span> on{' '}
        <span className="font-mono">succeeded</span>,{' '}
        <span className="font-mono">failed</span>, and{' '}
        <span className="font-mono">cancelled</span> events. Omitted on{' '}
        <span className="font-mono">started</span>.
      </>
    ),
  },
  {
    field: 'data.outcome.error',
    type: 'string',
    presence: 'on failed only',
    description:
      'Human-readable error message. Omitted when the workflow / step succeeded or was cancelled.',
  },
  {
    field: 'data.outcome.duration_ms',
    type: 'integer',
    presence: 'on terminal events',
    description: 'How long the workflow / step took to run, in milliseconds.',
  },
]

const LINK_FIELDS: FieldRow[] = [
  {
    field: 'data.links.org',
    type: 'string (url)',
    presence: 'always',
    description: 'Dashboard URL for the org.',
  },
  {
    field: 'data.links.install',
    type: 'string (url)',
    presence: 'when applicable',
    description: 'Dashboard URL for the install. Set when the workflow owner is an install.',
  },
  {
    field: 'data.links.workflow',
    type: 'string (url)',
    presence: 'when applicable',
    description: 'Dashboard URL for the workflow run.',
  },
  {
    field: 'data.links.sandbox',
    type: 'string (url)',
    presence: 'when applicable',
    description: 'Dashboard URL for the sandbox.',
  },
  {
    field: 'data.links.component',
    type: 'string (url)',
    presence: 'when applicable',
    description: 'Dashboard URL for the component.',
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
        The webhook surface exposes two primitives:{' '}
        <span className="font-mono">com.nuon.workflow.lifecycle.v1</span> for
        workflow-level events and{' '}
        <span className="font-mono">com.nuon.workflow_step.lifecycle.v1</span>{' '}
        for step-level events. Both share the same envelope and{' '}
        <span className="font-mono">data</span> shape, with{' '}
        <span className="font-mono">data.step</span> present only on step
        events. Each workflow / step emits a{' '}
        <span className="font-mono">started</span> transition when it begins
        and a terminal{' '}
        <span className="font-mono">succeeded</span>,{' '}
        <span className="font-mono">failed</span>, or{' '}
        <span className="font-mono">cancelled</span> transition when it ends.
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
        Describes the kind of event, the transition, and the org it belongs to.
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
        data.workflow
      </Text>
      <Text variant="subtext" theme="neutral">
        Identifies the workflow this event is about. Present on every event.
      </Text>
      <PropertyGrid
        className="rounded-md border p-4"
        columns={FIELD_COLUMNS}
        values={WORKFLOW_FIELDS}
        gridTemplate={GRID_TEMPLATE}
      />
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        data.step
      </Text>
      <Text variant="subtext" theme="neutral">
        Present only on{' '}
        <span className="font-mono">workflow_step</span> events. Identifies the
        specific step within the workflow and the resource it manipulates.
      </Text>
      <PropertyGrid
        className="rounded-md border p-4"
        columns={FIELD_COLUMNS}
        values={STEP_FIELDS}
        gridTemplate={GRID_TEMPLATE}
      />
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        data.parent
      </Text>
      <Text variant="subtext" theme="neutral">
        Present when this workflow was launched from another workflow's step
        (for example, an action workflow run launched from a deploy step).
        Omitted for top-level workflows.
      </Text>
      <PropertyGrid
        className="rounded-md border p-4"
        columns={FIELD_COLUMNS}
        values={PARENT_FIELDS}
        gridTemplate={GRID_TEMPLATE}
      />
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        data.outcome
      </Text>
      <Text variant="subtext" theme="neutral">
        Set on terminal transitions (<span className="font-mono">succeeded</span>,{' '}
        <span className="font-mono">failed</span>,{' '}
        <span className="font-mono">cancelled</span>). Omitted on{' '}
        <span className="font-mono">started</span>.
      </Text>
      <PropertyGrid
        className="rounded-md border p-4"
        columns={FIELD_COLUMNS}
        values={OUTCOME_FIELDS}
        gridTemplate={GRID_TEMPLATE}
      />
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        data.links
      </Text>
      <Text variant="subtext" theme="neutral">
        Dashboard URLs for the entities referenced in the event. Only links
        that can be resolved are included.
      </Text>
      <PropertyGrid
        className="rounded-md border p-4"
        columns={FIELD_COLUMNS}
        values={LINK_FIELDS}
        gridTemplate={GRID_TEMPLATE}
      />
    </div>
  </div>
)
