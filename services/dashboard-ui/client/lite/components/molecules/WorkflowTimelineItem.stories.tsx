import type { ReactNode } from 'react'
import type { TWorkflow } from '@/types/ctl-api.types'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { WorkflowTimelineItem } from './WorkflowTimelineItem'

export default {
  title: 'lite/molecules/WorkflowTimelineItem',
}

const List = ({ children }: { children: ReactNode }) => (
  <ol className="flex min-w-0 flex-col">{children}</ol>
)

const run = (workflow: Partial<TWorkflow>): TWorkflow =>
  ({
    id: 'wflq7fplr1up5atx5zpxotbabm',
    owner_type: 'app_branches',
    type: 'app_branches_manual_update',
    name: 'Run',
    created_at: '2026-09-10T12:00:00Z',
    status: { status: 'success' },
    ...workflow,
  }) as TWorkflow

export const Overview = () => (
  <ComponentDocs
    name="WorkflowTimelineItem"
    tier="molecule"
    summary="One workflow run in an activity timeline, with its status, identity and every marker the run carries."
    use={[
      'Render a row inside WorkflowTimeline, for either an install workflow or an app branch run.',
      'Pass pendingApprovalIds when an approvals source is available, so the pending-approval marker renders.',
      'Pass driftedWorkflowIds so a drift scan that found drift is distinguishable from one that did not.',
    ]}
    avoid={[
      'Do not use it for a single run outside a timeline; it renders an li and expects a list.',
      'Do not reimplement the status chip. It comes from the shared workflow badge map.',
      'Do not pass an href unless the target route exists; an unlinked title is correct when it does not.',
    ]}
    rules={[
      'There is one status chip and it always sits beside the title, never right-aligned.',
      'The chip label and theme come from WORKFLOW_BADGE_MAP, falling back to an in-flight label for statuses the map does not cover.',
      'Preview markers only render for branch runs; drift markers only render for install workflows.',
      'A preview run says preview once. Plan-only is not badged separately because preview already implies it.',
      'A service-account trigger renders as "a service account" with the email in a tooltip.',
      'Duration renders only once the run has finished.',
      'loading keeps the marker and spacing and replaces text with skeletons.',
    ]}
    props={[
      {
        name: 'workflow',
        type: 'TWorkflow',
        description: 'The run to render.',
      },
      {
        name: 'href',
        type: 'string',
        description:
          'Makes the title a link. Omit when no detail route exists.',
      },
      {
        name: 'pendingApprovalIds',
        type: 'ReadonlySet<string>',
        description:
          'Workflow ids with approvals outstanding, driving the pending-approval marker.',
      },
      {
        name: 'driftedWorkflowIds',
        type: 'ReadonlySet<string>',
        description:
          'Workflow ids that reported drifted objects, driving the drift-detected marker.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders the skeleton row.',
      },
    ]}
  />
)

export const BranchRun = () => (
  <List>
    <WorkflowTimelineItem
      workflow={run({
        name: 'PR #128',
        status: { status: 'success' },
        started_at: '2026-09-10T12:00:00Z',
        finished_at: '2026-09-10T12:04:20Z',
        app_branch_runs: [
          {
            event_type: 'pull_request',
            pr_number: 128,
            plan_only: true,
            preview: {
              source: 'pr',
              mode: 'plan-only',
              install_name: 'acme-staging',
            },
          },
        ],
      } as Partial<TWorkflow>)}
    />
    <WorkflowTimelineItem
      workflow={run({
        id: 'wflq7fplr1up5atx5zpxotbabn',
        name: 'VCS push',
        type: 'app_branches_config_repo_update',
        created_at: '2026-09-09T09:30:00Z',
        status: { status: 'in-progress' },
      })}
    />
    <WorkflowTimelineItem
      workflow={run({
        id: 'wflq7fplr1up5atx5zpxotbabo',
        name: 'Run',
        created_at: '2026-09-08T18:05:00Z',
        status: { status: 'error' },
        started_at: '2026-09-08T18:05:00Z',
        finished_at: '2026-09-08T18:06:02Z',
        app_branch_runs: [{ event_type: 'manual' }],
      } as Partial<TWorkflow>)}
    />
  </List>
)

export const InstallWorkflow = () => (
  <List>
    <WorkflowTimelineItem
      workflow={run({
        id: 'wflq7fplr1up5atx5zpxotbabp',
        owner_type: 'installs',
        type: 'drift_run',
        name: 'Deploying to install (api)',
        plan_only: true,
        created_at: '2026-09-10T11:00:00Z',
        status: { status: 'success' },
      })}
      driftedWorkflowIds={new Set(['wflq7fplr1up5atx5zpxotbabp'])}
    />
    <WorkflowTimelineItem
      workflow={run({
        id: 'wflq7fplr1up5atx5zpxotbabq',
        owner_type: 'installs',
        type: 'manual_deploy',
        name: 'Deploying to install (web)',
        created_at: '2026-09-09T11:00:00Z',
        status: { status: 'approval-awaiting' },
        approval_option: 'prompt',
      })}
    />
    <WorkflowTimelineItem
      workflow={run({
        id: 'wflq7fplr1up5atx5zpxotbabr',
        owner_type: 'installs',
        type: 'deploy_components',
        name: 'Deploying all components',
        created_at: '2026-09-08T11:00:00Z',
        status: { status: 'success' },
        approval_option: 'approve-all',
        metadata: { approval_type: 'install-config' },
      } as Partial<TWorkflow>)}
    />
  </List>
)

export const PendingApproval = () => (
  <List>
    <WorkflowTimelineItem
      workflow={run({
        name: 'Run',
        approval_option: 'prompt',
        status: { status: 'in-progress' },
      })}
      pendingApprovalIds={new Set(['wflq7fplr1up5atx5zpxotbabm'])}
    />
  </List>
)

export const ServiceAccountTrigger = () => (
  <List>
    <WorkflowTimelineItem
      workflow={run({
        created_by: { email: 'runner@serviceaccount.nuon.co' },
      } as Partial<TWorkflow>)}
    />
  </List>
)

export const Linked = () => (
  <List>
    <WorkflowTimelineItem workflow={run({})} href="#run" />
  </List>
)

export const LongNames = () => (
  <List>
    <WorkflowTimelineItem
      workflow={run({
        name: 'Deploying to install (rds_cluster_temporal_with_a_very_long_component_name)',
        app_branch_runs: [
          {
            event_type: 'pull_request',
            pr_number: 4821,
            preview: {
              source: 'pr',
              mode: 'apply',
              install_name: 'acme-production-eu-west-1',
            },
          },
        ],
      } as Partial<TWorkflow>)}
    />
  </List>
)

export const Loading = () => (
  <List>
    <WorkflowTimelineItem workflow={run({})} loading />
    <WorkflowTimelineItem workflow={run({})} loading />
  </List>
)
