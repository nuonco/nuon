import { useState } from 'react'
import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { getStepKind } from '@/utils/workflow-utils'
import { WorkflowStepGroup } from './WorkflowStepGroup'
import { WorkflowStepRoundGroup } from './WorkflowStepRoundGroup'
import { WorkflowStepRow } from './WorkflowStepRow'

export interface IWorkflowSteps {
  approvalPrompt?: boolean
  planOnly?: boolean
  workflowSteps: TWorkflowStep[]
  eagerStepsLoaded?: boolean
  allStepsLoaded?: boolean
}

export const WorkflowSteps = ({
  approvalPrompt = false,
  planOnly = false,
  workflowSteps,
  eagerStepsLoaded = false,
  allStepsLoaded = false,
}: IWorkflowSteps) => {
  const [searchName, setSearchName] = useState<string>('')

  const filteredSteps = workflowSteps
    .filter((step) => step.execution_type !== 'hidden')
    .filter((step) => step.name.includes(searchName))

  // Steps are organized by `group_idx` (plan+apply for one component share a
  // group; independent steps get their own). A group is retried two ways:
  //   • group retry — apply fails and the whole group re-runs, cloning every
  //     step with an incremented `group_retry_idx`. These render as one card
  //     with the current round on top and prior rounds nested (WorkflowStepRoundGroup).
  //   • step retry — a single step auto-retries on its own (e.g. a plan), keeping
  //     `group_retry_idx` at 0 while appending same-kind attempts. These collapse
  //     per kind (WorkflowStepGroup), with single-attempt kinds as plain rows.
  const groupsByGroupIdx = new Map<string, TWorkflowStep[]>()
  let soloIdx = 0
  for (const step of filteredSteps) {
    const key = step.group_idx != null ? `g${step.group_idx}` : `s${soloIdx++}`
    const group = groupsByGroupIdx.get(key)
    if (group) {
      group.push(step)
    } else {
      groupsByGroupIdx.set(key, [step])
    }
  }

  const renderGroup = (groupSteps: TWorkflowStep[]) => {
    const rounds = new Set(groupSteps.map((s) => s.group_retry_idx ?? 0))
    if (rounds.size > 1) {
      return (
        <WorkflowStepRoundGroup
          key={`round-${groupSteps[0].id}`}
          steps={groupSteps}
          approvalPrompt={approvalPrompt}
          planOnly={planOnly}
        />
      )
    }

    const byKind = new Map<string, TWorkflowStep[]>()
    for (const step of groupSteps) {
      const kind = getStepKind(step)
      const kindSteps = byKind.get(kind)
      if (kindSteps) {
        kindSteps.push(step)
      } else {
        byKind.set(kind, [step])
      }
    }

    return [...byKind.values()].map((kindSteps) => {
      if (kindSteps.length > 1) {
        return (
          <WorkflowStepGroup
            key={kindSteps[0].id}
            steps={kindSteps}
            approvalPrompt={approvalPrompt}
            planOnly={planOnly}
          />
        )
      }

      const step = kindSteps[0]
      return (
        <div key={step.id} className="flex px-4 py-2">
          <WorkflowStepRow
            step={step}
            approvalPrompt={approvalPrompt}
            planOnly={planOnly}
            showRetry
          />
        </div>
      )
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <SearchInput
        placeholder="Search workflow steps"
        value={searchName}
        onChange={setSearchName}
      />
      {filteredSteps.length ? (
        <div className="flex flex-col border rounded-md divide-y overflow-hidden">
          {[...groupsByGroupIdx.values()].flatMap((group) => renderGroup(group))}

          {eagerStepsLoaded && !allStepsLoaded ? (
            <div className="flex items-center gap-4 px-4 py-2">
              <span className="flex items-center justify-center w-6 h-6 shrink-0">
                <Loading className="h-5 w-5 text-cool-grey-500 dark:text-white/50" />
              </span>
              <Text variant="base" theme="neutral">
                Generating more steps…
              </Text>
            </div>
          ) : null}
        </div>
      ) : (
        <EmptyState
          variant="table"
          emptyMessage={
            workflowSteps.length
              ? 'No workflow steps match your search. Try adjusting your search criteria.'
              : 'Steps will appear here once the workflow has been generated.'
          }
          emptyTitle={workflowSteps.length ? 'No steps found' : 'Workflow steps not available'}
        />
      )}
    </div>
  )
}

const SKELETON_TITLE_WIDTHS = ['12rem', '16rem', '10rem', '14rem', '11rem']

export const WorkflowStepsSkeleton = () => {
  return (
    <div className="flex flex-col border rounded-md divide-y overflow-hidden">
      {SKELETON_TITLE_WIDTHS.map((width, idx) => (
        <div key={idx} className="flex items-center gap-4 px-4 py-2">
          <span className="flex items-center justify-center w-6 h-6 shrink-0">
            <Loading className="h-5 w-5 text-cool-grey-500 dark:text-white/50" />
          </span>
          <Skeleton height="1rem" width={width} />
        </div>
      ))}
    </div>
  )
}
