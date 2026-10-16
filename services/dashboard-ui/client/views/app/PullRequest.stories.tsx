export default {
  title: 'Views / Apps / Pull request',
  meta: { fullBleed: true, installViews: true },
}

import { PageTitle } from '@/components/navigation/PageTitle'
import { PullRequestScope } from '@/components/playground/pull-request/PullRequestScope'
import {
  configuredPullRequest,
  firstRunOnly,
  needsInputPullRequest,
  previewRuns,
  type TPlaygroundPullRequest,
  type TPlaygroundPullRequestRun,
} from '@/components/playground/pull-request/fixtures'
import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const BRANCH_ROUTE = ':orgId/apps/:appId/branches/:branchId'

const page = (
  pullRequest: TPlaygroundPullRequest,
  runs: TPlaygroundPullRequestRun[]
) => (
  <InstallView
    fixture={appFixture('picker')}
    path={appPath('/branches/br-1')}
    routeElements={{
      [BRANCH_ROUTE]: (
        <>
          <PageTitle segments={[`#${pullRequest.number}`, 'Payments']} />
          <PullRequestScope
            pullRequest={pullRequest}
            runs={runs}
            initialStep="detail"
          />
        </>
      ),
    }}
  />
)

export const WithHistory = () => page(configuredPullRequest, previewRuns)
export const NeedsAnInstall = () => page(needsInputPullRequest, firstRunOnly)
