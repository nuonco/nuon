export default {
  title: 'Features / Orgs / Recent updates',
}

import type { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { UpdateCard } from '@/components/orgs/BranchActivityFeed'
import { RecentUpdates } from './RecentUpdates'
import {
  ALL_UPDATES,
  AWAITING_APPROVAL,
  BRANCH_PLAN_APPROVAL,
  CANCELLED,
  FAILED,
  FAILED_BEFORE_ROLLOUT,
  LABEL_SELECTOR_NO_MATCHES,
  LONG_NAMES,
  MANUAL_NO_COMMIT,
  MANY_GROUPS,
  NO_INSTALL_GROUPS,
  PENDING,
  RUNNING,
  SUCCEEDED,
  toStoryItems,
} from './RecentUpdates.fixtures'

const Frame = ({ children }: { children: ReactNode }) => (
  <div className="max-w-3xl p-6 flex flex-col gap-4">{children}</div>
)

export const MixedFeed = () => (
  <Frame>
    <RecentUpdates items={toStoryItems(ALL_UPDATES)} />
  </Frame>
)

export const AwaitingApproval = () => (
  <Frame>
    <RecentUpdates
      items={toStoryItems(ALL_UPDATES)}
      initialFilter="awaiting-approval"
    />
  </Frame>
)

export const Failed = () => (
  <Frame>
    <RecentUpdates items={toStoryItems(ALL_UPDATES)} initialFilter="failed" />
  </Frame>
)

export const InProgress = () => (
  <Frame>
    <RecentUpdates
      items={toStoryItems(ALL_UPDATES)}
      initialFilter="in-progress"
    />
  </Frame>
)

export const AllSucceeded = () => (
  <Frame>
    <RecentUpdates
      items={toStoryItems([
        SUCCEEDED,
        MANUAL_NO_COMMIT,
        LABEL_SELECTOR_NO_MATCHES,
        NO_INSTALL_GROUPS,
      ])}
    />
  </Frame>
)

export const NoMatchesForFilter = () => (
  <Frame>
    <RecentUpdates
      items={toStoryItems([SUCCEEDED, NO_INSTALL_GROUPS])}
      initialFilter="failed"
    />
  </Frame>
)

export const Empty = () => (
  <Frame>
    <RecentUpdates items={[]} />
  </Frame>
)

export const Loading = () => (
  <Frame>
    <RecentUpdates items={[]} isLoading />
  </Frame>
)

const CARD_CASES = [
  ['Pending, rollout not started', PENDING],
  ['Running, installs in progress and queued', RUNNING],
  ['Awaiting install approvals', AWAITING_APPROVAL],
  ['Awaiting branch plan approval', BRANCH_PLAN_APPROVAL],
  ['Running across more groups than fit', MANY_GROUPS],
  ['Failed during rollout', FAILED],
  ['Failed before rollout', FAILED_BEFORE_ROLLOUT],
  ['Succeeded across three groups', SUCCEEDED],
  ['Cancelled mid-rollout', CANCELLED],
  ['Manual run without commit metadata', MANUAL_NO_COMMIT],
  ['Label selector matched no installs', LABEL_SELECTOR_NO_MATCHES],
  ['No install groups configured', NO_INSTALL_GROUPS],
  ['Long app, branch, and commit text', LONG_NAMES],
] as const

export const CardVariants = () => (
  <div className="max-w-3xl p-6 flex flex-col gap-6">
    {CARD_CASES.map(([label, update]) => {
      const [item] = toStoryItems([update])
      return (
        <section key={update.runId} className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            {label}
          </Text>
          {item ? <UpdateCard item={item} /> : null}
        </section>
      )
    })}
  </div>
)
