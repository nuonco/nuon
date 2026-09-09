import { DateTime } from 'luxon'
import type { TInstall } from '@/types/ctl-api.types'
import { Badge } from '../../atoms/Badge'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { InstallsTable, type IInstallFilter } from './InstallsTable'

export default {
  title: 'lite/organisms/InstallsTable',
}

const INSTALLS: TInstall[] = [
  {
    id: 'inst_payments_production',
    name: 'Production',
    app_id: 'app_payments',
    app: { id: 'app_payments', name: 'Payments API' },
    runner_status: 'active',
    sandbox_status: 'active',
    composite_component_status: 'active',
    composite_health_status: 'healthy',
    drifted_objects: [],
    cloud_platform: 'aws',
    aws_account: { region: 'us-west-2' },
    app_branch: { id: 'branch_main', name: 'main' },
    labels: { env: 'production', tier: 'critical' },
    updated_at: DateTime.now().minus({ minutes: 3 }).toISO(),
  },
  {
    id: 'inst_payments_staging',
    name: 'Staging',
    app_id: 'app_payments',
    app: { id: 'app_payments', name: 'Payments API' },
    runner_status: 'active',
    sandbox_status: 'queued',
    composite_component_status: 'deploying',
    cloud_platform: 'gcp',
    gcp_account: { region: 'us-central1' },
    app_branch: { id: 'branch_next', name: 'next' },
    labels: { env: 'staging' },
    updated_at: DateTime.now().minus({ minutes: 18 }).toISO(),
  },
  {
    id: 'inst_dashboard_preview',
    name: 'Preview',
    app_id: 'app_dashboard',
    app: { id: 'app_dashboard', name: 'Dashboard' },
    runner_status: 'offline',
    sandbox_status: 'active',
    sandbox_health_status: 'unhealthy',
    sandbox_health_message: 'The sandbox cluster stopped reporting node health.',
    composite_component_status: 'failed',
    composite_component_status_description:
      'Terraform apply exited with status 1.',
    composite_health_status: 'degraded',
    drifted_objects: [{ target_id: 'cmp_api' }, { target_id: 'cmp_worker' }],
    cloud_platform: 'azure',
    azure_account: { location: 'westeurope' },
    labels: {},
    updated_at: DateTime.now().minus({ hours: 2 }).toISO(),
  },
  {
    id: 'inst_dashboard_teardown',
    name: 'Old preview',
    app_id: 'app_dashboard',
    app: { id: 'app_dashboard', name: 'Dashboard' },
    runner_status: 'active',
    sandbox_status: 'active',
    composite_component_status: 'active',
    lifecycle_phase: { phase: 'deprovisioned' },
    cloud_platform: 'aws',
    aws_account: { region: 'us-east-1' },
    labels: {},
    updated_at: DateTime.now().minus({ days: 3 }).toISO(),
  },
]

const FILTER_LABEL_COLORS: Record<string, string> = {
  env: '#2563eb',
  tier: '#dc2626',
}

const labelOption = (value: string) => {
  const [key, ...rest] = value.split(':')
  return {
    value,
    textValue: value,
    label: (
      <Badge
        variant="code"
        labelKey={key}
        labelValue={rest.join(':')}
        color={FILTER_LABEL_COLORS[key]}
      />
    ),
  }
}

const labels: IInstallFilter = {
  label: 'Labels',
  options: [
    labelOption('env:production'),
    labelOption('env:staging'),
    labelOption('tier:critical'),
  ],
  selected: new Set(),
  onToggle: () => {},
  onIsolate: () => {},
  onReset: () => {},
  constrained: false,
}

const branches: IInstallFilter = {
  label: 'Branches',
  options: [
    { value: '__none__', label: 'No branch' },
    { value: 'main', label: 'main' },
    { value: 'next', label: 'next' },
  ],
  selected: new Set(),
  onToggle: () => {},
  onIsolate: () => {},
  onReset: () => {},
  constrained: false,
}

const props = {
  orgId: 'org_example',
  search: '',
  onSearchChange: () => {},
  offset: 0,
  pageSize: 20,
  hasNext: false,
  onOffsetChange: () => {},
  labelFilter: labels,
  branchFilter: branches,
  labelColors: {
    app_payments: { env: '#2563eb', tier: '#dc2626' },
  },
}

export const Overview = () => (
  <ComponentDocs
    name="InstallsTable"
    tier="organism"
    summary="The org's installs rendered through Table, with search, label and branch filters, and pagination bound to /v1/installs."
    use={[
      'Render it from the Installs page through InstallsTableContainer.',
      'Pass label and branch filters as IInstallFilter so the dropdowns stay URL-backed.',
    ]}
    avoid={[
      'Do not fetch installs or filter options inside the presentation component.',
      'Do not filter the resolved page in the browser — the API owns narrowing.',
      'Do not collapse the status axes into one install status — an install has no holistic status.',
    ]}
    rules={[
      'Every rendered label is a Badge — in the table cell, the card, and the filter options alike.',
      'Label colors come from the owning app, so an install with no app_id match falls back to the neutral badge.',
      'Runner, sandbox, and components are independent axes, each its own marker keyed to its own icon.',
      'Health and drift render only once the API reports them; health stays empty until the evaluator runs.',
      'A deprovisioning or deprovisioned lifecycle phase overrides axes still reading active, so a torn-down install never looks live.',
      'Sandbox health replaces the sandbox status only while the sandbox itself is active.',
      'Stack status is absent from /v1/installs and belongs on the install page, not this table.',
      'An empty selected set means the filter is off, so every install passes.',
      'A constrained filter reports its active count on the dropdown trigger.',
      'Search and every filter change reset the offset in the container.',
      'Pagination sits below Table, not in the toolbar.',
    ]}
    props={[
      {
        name: 'installs',
        type: 'TInstall[]',
        description: 'Resolved installs for the current page.',
      },
      {
        name: 'orgId',
        type: 'string',
        description: 'Org segment used to build install and app hrefs.',
      },
      {
        name: 'search',
        type: 'string',
        description: 'Current search term from the URL.',
      },
      {
        name: 'onSearchChange',
        type: '(value: string) => void',
        description: 'Writes the search term and resets the offset.',
      },
      {
        name: 'labelFilter',
        type: 'IInstallFilter',
        description:
          'Label key/value options, selection, and reset handlers. Options carry a colored Badge as their label, matching the row badges.',
      },
      {
        name: 'labelColors',
        type: 'Record<string, Record<string, string>>',
        description: 'Label key colors per app id, from /v1/apps/:id/labels.',
      },
      {
        name: 'branchFilter',
        type: 'IInstallFilter',
        description: 'App branch options, including the no-branch sentinel.',
      },
      {
        name: 'offset',
        type: 'number',
        description: 'Row offset of the current page.',
      },
      {
        name: 'pageSize',
        type: 'number',
        description: 'Rows requested per page.',
      },
      {
        name: 'hasNext',
        type: 'boolean',
        description: 'Whether a full page came back, enabling next.',
      },
      {
        name: 'onOffsetChange',
        type: '(offset: number) => void',
        description: 'Moves the page window.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders Table skeletons on the cold load.',
      },
      {
        name: 'fetching',
        type: 'boolean',
        default: 'false',
        description: 'Disables pagination while a page is in flight.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Swaps the empty state to failure copy.',
      },
    ]}
  />
)

export const Default = () => <InstallsTable {...props} installs={INSTALLS} />

export const Loading = () => <InstallsTable {...props} installs={[]} loading />

export const Empty = () => <InstallsTable {...props} installs={[]} />

export const Filtered = () => (
  <InstallsTable
    {...props}
    installs={INSTALLS.slice(0, 1)}
    labelFilter={{
      ...labels,
      selected: new Set(['env:production']),
      constrained: true,
    }}
  />
)

export const ErrorState = () => (
  <InstallsTable {...props} installs={[]} error={new Error('Request failed')} />
)
