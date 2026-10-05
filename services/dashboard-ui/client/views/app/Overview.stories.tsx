export default {
  title: 'Views / Apps / Overview',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1')} />
)

export const Succeeded = () => page('overview-succeeded')
export const LongCommit = () => page('overview-long-commit')
export const RollingOut = () => page('overview-rolling-out')
export const AwaitingApproval = () => page('overview-awaiting')
export const Failed = () => page('overview-failed')
export const NoRuns = () => page('overview-no-runs')
export const NoPlan = () => page('overview-no-plan')
export const Loading = () => page('overview-loading')
