import { WorkflowTimelineComponent } from '@/components/workflows/WorkflowTimeline'
import { WorkflowFilters } from '@/components/workflows/filters/WorkflowFilters'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { WorkflowApprovalsContext } from '@/providers/workflow-approvals-provider'
import type { TWorkflow } from '@/types'
import type { TLatestRollout } from './fixtures'

const TRIGGER: Record<TLatestRollout['source']['kind'], string> = {
  'pull-request': 'pull_request',
  tag: 'tag',
  manual: 'manual',
  commit: 'push',
}

const toWorkflow = (rollout: TLatestRollout): TWorkflow => {
  const { source, commit } = rollout
  const prNumber = source.kind === 'pull-request' ? source.number : undefined
  return {
    id: rollout.id,
    name: 'Run',
    type: 'app_branch_run',
    owner_type: 'app_branches',
    finished: true,
    plan_only: false,
    status: { status: rollout.status },
    created_at: commit.createdAt,
    updated_at: commit.createdAt,
    finished_at: commit.createdAt,
    created_by: { email: commit.author },
    app_branch_runs: [
      {
        head_sha: commit.sha,
        pr_number: prNumber,
        status: rollout.status,
        metadata: {
          trigger: TRIGGER[source.kind],
          pr_number: prNumber,
          tag: source.kind === 'tag' ? source.tag : undefined,
        },
        vcs_connection_commit: {
          sha: commit.sha,
          message: commit.message,
          author_name: commit.author,
          created_at: commit.createdAt,
        },
      },
    ],
    steps: rollout.installGroups.map((group) => ({
      name: `Deploy install group: ${group.name}`,
      status: {
        status: group.status,
        metadata: {
          installs: group.installs.map((install) => ({
            install_id: install.id,
            status: install.status,
          })),
        },
      },
    })),
  } as unknown as TWorkflow
}

const approvals = { approvals: [], isLoading: false, refresh: () => {} }

export const PreviousRunsPage = ({
  rollouts,
  basePath,
}: {
  rollouts: TLatestRollout[]
  basePath: string
}) => (
  <WorkflowApprovalsContext.Provider value={approvals}>
    <PageSection>
      <SectionHeader
        title="Previous runs"
        description="Every run from this branch, newest first."
      />
      <WorkflowFilters owner="app" />
      <WorkflowTimelineComponent
        workflows={rollouts.map(toWorkflow)}
        pagination={{ hasNext: false, offset: 0, limit: 20 }}
        orgId="org-1"
        getWorkflowHref={(run) => `${basePath}/runs/${run.id}`}
      />
    </PageSection>
  </WorkflowApprovalsContext.Provider>
)
