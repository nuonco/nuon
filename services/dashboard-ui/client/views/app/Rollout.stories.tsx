export default {
  title: 'Views / Apps / Rollout',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/rollout')} />
)

export const Succeeded = () => page('rollout-succeeded')
export const RollingOut = () => page('rollout-rolling-out')
export const AwaitingApproval = () => page('rollout-awaiting')
export const Failed = () => page('rollout-failed')
export const NoRuns = () => page('rollout-no-runs')
export const NoPlan = () => page('rollout-no-plan')
export const Loading = () => page('rollout-loading')
