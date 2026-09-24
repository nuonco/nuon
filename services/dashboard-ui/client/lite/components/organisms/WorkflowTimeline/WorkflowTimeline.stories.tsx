import type { TWorkflow } from '@/types/ctl-api.types'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import {
  WORKFLOW_DATE_LABELS,
  WORKFLOW_PREVIEW_LABELS,
  WORKFLOW_STATUS_LABELS,
  WORKFLOW_TYPE_LABELS,
  workflowStatusOptions,
  workflowTypeOptions,
} from '../../../utils/workflow-filters'
import { WorkflowTimeline, type IWorkflowFilter } from './WorkflowTimeline'

export default {
  title: 'lite/organisms/WorkflowTimeline',
}

const filter = (
  label: string,
  options: readonly { value: string; label: string }[]
): IWorkflowFilter<string> => ({
  label,
  options,
  selected: new Set<string>(),
  onToggle: () => {},
  onIsolate: () => {},
  onReset: () => {},
})

const statusFilter = filter(
  'Status',
  workflowStatusOptions().map((value) => ({
    value,
    label: WORKFLOW_STATUS_LABELS[value],
  }))
)

const typeFilter = filter(
  'Type',
  workflowTypeOptions('app').map((value) => ({
    value,
    label: WORKFLOW_TYPE_LABELS[value],
  }))
)

const previewFilter = filter(
  'Preview',
  (['preview', 'rollout'] as const).map((value) => ({
    value,
    label: WORKFLOW_PREVIEW_LABELS[value],
  }))
)

const dateFilter = filter(
  'Date',
  (['24h', '7d', '30d'] as const).map((value) => ({
    value,
    label: WORKFLOW_DATE_LABELS[value],
  }))
)

const run = (workflow: Partial<TWorkflow>): TWorkflow =>
  ({
    owner_type: 'app_branches',
    type: 'app_branches_manual_update',
    name: 'Run',
    status: { status: 'success' },
    ...workflow,
  }) as TWorkflow

const runs: TWorkflow[] = [
  run({
    id: 'wflq7fplr1up5atx5zpxotbab1',
    name: 'PR #128',
    created_at: '2026-09-10T12:00:00Z',
    started_at: '2026-09-10T12:00:00Z',
    finished_at: '2026-09-10T12:04:20Z',
  }),
  run({
    id: 'wflq7fplr1up5atx5zpxotbab2',
    name: 'VCS push',
    type: 'app_branches_config_repo_update',
    created_at: '2026-09-09T09:30:00Z',
    status: { status: 'in-progress' },
  }),
  run({
    id: 'wflq7fplr1up5atx5zpxotbab3',
    name: 'Run',
    created_at: '2026-09-08T18:05:00Z',
    status: { status: 'error' },
  }),
]

const base = {
  search: '',
  onSearchChange: () => {},
  offset: 0,
  pageSize: 20,
  hasNext: false,
  onOffsetChange: () => {},
  statusFilter,
  typeFilter,
  previewFilter,
  dateFilter,
}

export const Overview = () => (
  <ComponentDocs
    name="WorkflowTimeline"
    tier="organism"
    summary="The activity timeline: every run on an install or an app branch, with server-side search, filters and pagination."
    use={[
      'Mount it on an activity page through its container, which owns the query, the SSE stream and the URL state.',
      'Pass the owner to the container so the type filter offers that owner\u2019s run types.',
      'Render it presentationally in stories and tests, with filter controls supplied as props.',
    ]}
    avoid={[
      'Do not add a second "active runs" block above it. Running is a status filter, not another component.',
      'Do not filter or paginate the rows in the page; every control resolves server-side.',
      'Do not build a date range picker here. The date axis is presets only.',
      'Do not build a three-state preview toggle. Selecting neither option means both.',
      'Do not offer the preview axis on an install. Plan only is fully determined by type there, and the Drift scan type option covers it.',
    ]}
    rules={[
      'Rows are grouped by day by Timeline, newest first.',
      'Empty, filtered-to-nothing and failed states render distinct content in the same slot.',
      'Search matches the run title or its ID, and every control round-trips through the URL.',
      'loading renders Timeline\u2019s own skeleton; fetching only dims pagination.',
    ]}
    props={[
      { name: 'workflows', type: 'TWorkflow[]', description: 'Runs for the current page.' },
      { name: 'search', type: 'string', description: 'Current search term.' },
      { name: 'onSearchChange', type: '(value: string) => void', description: 'Called as the search term changes.' },
      { name: 'offset', type: 'number', description: 'Current page offset.' },
      { name: 'pageSize', type: 'number', description: 'Rows per page.' },
      { name: 'hasNext', type: 'boolean', description: 'Whether a further page exists.' },
      { name: 'onOffsetChange', type: '(offset: number) => void', description: 'Called when the page changes.' },
      { name: 'statusFilter', type: 'IWorkflowFilter<string>', description: 'Status axis control.' },
      { name: 'typeFilter', type: 'IWorkflowFilter<string>', description: 'Run type axis control.' },
      { name: 'previewFilter', type: 'IWorkflowFilter<string>', description: 'Preview versus rollout axis control. Omitted on an install, where the Drift scan type option already expresses it.' },
      { name: 'dateFilter', type: 'IWorkflowFilter<string>', description: 'Date preset axis control.' },
      { name: 'getWorkflowHref', type: '(workflow: TWorkflow) => string | undefined', description: 'Row link target. Rows render unlinked when it is omitted.' },
      { name: 'pendingApprovalIds', type: 'ReadonlySet<string>', description: 'Runs with approvals outstanding.' },
      { name: 'driftedWorkflowIds', type: 'ReadonlySet<string>', description: 'Runs that reported drifted objects.' },
      { name: 'filtered', type: 'boolean', default: 'false', description: 'Switches the empty state to the filtered-to-nothing wording.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'First load, no rows yet.' },
      { name: 'fetching', type: 'boolean', default: 'false', description: 'Revalidating with rows on screen.' },
      { name: 'error', type: 'unknown', description: 'Query error, switching the empty state to failure wording.' },
    ]}
  />
)

export const Default = () => <WorkflowTimeline {...base} workflows={runs} />

export const SingleDay = () => (
  <WorkflowTimeline {...base} workflows={[runs[0]]} />
)

export const Paginated = () => (
  <WorkflowTimeline {...base} workflows={runs} hasNext />
)

export const Loading = () => <WorkflowTimeline {...base} workflows={[]} loading />

export const Empty = () => <WorkflowTimeline {...base} workflows={[]} />

export const NoMatches = () => (
  <WorkflowTimeline {...base} workflows={[]} search="nothing" filtered />
)

export const FailedToLoad = () => (
  <WorkflowTimeline {...base} workflows={[]} error={new Error('boom')} />
)
