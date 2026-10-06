export default {
  title: 'Views / Dashboard',
  meta: { fullBleed: true, installViews: true },
}

import {
  ALL_UPDATES,
  AWAITING_APPROVAL,
  BRANCH_PLAN_APPROVAL,
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
} from '@/components/orgs/RecentUpdates/RecentUpdates.fixtures'
import type { TRecentUpdatePayloadItem } from '@/components/orgs/RecentUpdates/map-recent-updates'
import { InstallView } from '@/views/install/InstallView'
import {
  VIEW_ORG_ID,
  type TFixture,
} from '@/views/install/install-fixture'
import {
  viewInstall,
  withChrome,
} from '@/views/install/install-fixtures'

const fixture = withChrome(viewInstall(), () => undefined)

const dashboard = (
  updates?: TRecentUpdatePayloadItem[],
  dashboardFixture: TFixture = fixture
) => (
  <InstallView
    fixture={dashboardFixture}
    path={`/${VIEW_ORG_ID}`}
    initialQueryData={
      updates
        ? [
            {
              queryKey: ['recent-updates', VIEW_ORG_ID],
              data: { updates },
            },
          ]
        : undefined
    }
  />
)

export const MixedUpdates = () => dashboard(ALL_UPDATES)

export const InProgress = () =>
  dashboard([PENDING, RUNNING, MANY_GROUPS])

export const AwaitingApprovals = () =>
  dashboard([AWAITING_APPROVAL, BRANCH_PLAN_APPROVAL])

export const Failures = () => dashboard([FAILED, FAILED_BEFORE_ROLLOUT])

export const AllSucceeded = () =>
  dashboard([
    SUCCEEDED,
    MANUAL_NO_COMMIT,
    LABEL_SELECTOR_NO_MATCHES,
    NO_INSTALL_GROUPS,
  ])

export const LongContent = () => dashboard([LONG_NAMES, MANY_GROUPS])

export const Empty = () => dashboard([])

export const Loading = () => dashboard()

const sandboxFixture: TFixture = (url, init) => {
  if (url.pathname.endsWith('/orgs/current')) {
    return {
      body: {
        id: VIEW_ORG_ID,
        name: 'Acme',
        sandbox_mode: true,
        features: { 'new-app-ia': true },
      },
    }
  }
  return fixture(url, init)
}

export const SandboxMode = () => dashboard([RUNNING, SUCCEEDED], sandboxFixture)
