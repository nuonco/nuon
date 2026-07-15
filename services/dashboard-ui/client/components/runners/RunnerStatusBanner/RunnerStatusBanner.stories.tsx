export default {
  title: 'Runners/RunnerStatusBanner',
}

import { RunnerStatusBanner } from './RunnerStatusBanner'

const mngWarning =
  'The management process is not running. Remote restart, upgrade, and shutdown are unavailable until it recovers.'

export const Default = () => <RunnerStatusBanner warnings={[mngWarning]} />

export const Multiple = () => (
  <RunnerStatusBanner
    warnings={[mngWarning, 'Runner version is behind the configured version.']}
  />
)

export const Empty = () => <RunnerStatusBanner warnings={[]} />
