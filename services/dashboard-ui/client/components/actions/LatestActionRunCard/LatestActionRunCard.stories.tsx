export default {
  title: 'Actions/LatestActionRunCard',
}

import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import type { TInstallActionRun } from '@/types'
import { LatestActionRunCard } from './LatestActionRunCard'

const run = {
  id: 'arun-1',
  created_at: '2026-09-22T14:00:00Z',
  execution_time: 90_000_000_000,
  triggered_by_type: 'manual',
  status_v2: {
    status: 'success',
    status_human_description: 'Action finished.',
  },
} as TInstallActionRun

const trigger = <ActionTriggerType triggerType="manual" />

export const Default = () => (
  <LatestActionRunCard
    run={run}
    trigger={trigger}
    href="/org-1/installs/inst-1/actions/act-1/runs/arun-1"
  />
)

export const Flush = () => (
  <LatestActionRunCard
    flush
    run={run}
    trigger={trigger}
    href="/org-1/installs/inst-1/actions/act-1/runs/arun-1"
  />
)

export const InProgress = () => (
  <LatestActionRunCard
    run={
      {
        ...run,
        execution_time: undefined,
        status_v2: {
          status: 'in-progress',
          status_human_description: 'Action is running.',
        },
      } as TInstallActionRun
    }
    trigger={<ActionTriggerType triggerType="cron" />}
    href="/org-1/installs/inst-1/actions/act-1/runs/arun-1"
  />
)

export const Failed = () => (
  <LatestActionRunCard
    run={
      {
        ...run,
        status_v2: {
          status: 'error',
          status_human_description: 'Action failed.',
        },
      } as TInstallActionRun
    }
    trigger={trigger}
    href="/org-1/installs/inst-1/actions/act-1/runs/arun-1"
  />
)

export const Empty = () => <LatestActionRunCard />

export const Loading = () => (
  <LatestActionRunCard isLoading run={run} trigger={trigger} />
)
