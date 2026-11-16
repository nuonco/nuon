export default {
  title: 'Features / Installs / Alignment / Live build',
}

import {
  alignmentCommit,
  liveCommit,
  manualOrigin,
  runOrigin,
  selectedCommit,
} from '@/components/__fixtures__/install-alignment'
import { LiveBuild } from './LiveBuild'

const appliedAt = new Date(Date.now() - 72 * 3_600_000).toISOString()

export const Deploy = () => (
  <LiveBuild
    appliedAt={appliedAt}
    origin={runOrigin(liveCommit)}
    originHref="#run"
    repo="acme/payments"
    deployHref="#deploy"
  />
)

export const TaggedRun = () => (
  <LiveBuild
    appliedAt={appliedAt}
    origin={runOrigin(selectedCommit)}
    originHref="#run"
    repo="acme/payments"
    deployHref="#deploy"
  />
)

export const PushRun = () => (
  <LiveBuild
    appliedAt={appliedAt}
    origin={runOrigin(alignmentCommit(3))}
    originHref="#run"
    repo="acme/payments"
    deployHref="#deploy"
  />
)

export const ManualDeploy = () => (
  <LiveBuild
    appliedAt={appliedAt}
    origin={manualOrigin}
    originHref="#workflow"
    repo="acme/payments"
    deployHref="#deploy"
  />
)

export const NoOrigin = () => (
  <LiveBuild appliedAt={appliedAt} deployHref="#deploy" />
)
