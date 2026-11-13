export default {
  title: 'Features / Installs / Alignment / Live build',
}

import { alignmentBuild } from '@/components/__fixtures__/install-alignment'
import { LiveBuild } from './LiveBuild'

const build = alignmentBuild(
  'bld-live',
  '4f1d9c2a7be03e5d8a6c1b4f7a9e3c5b2d8f1a0e',
  76
)
const appliedAt = new Date(Date.now() - 72 * 3_600_000).toISOString()

export const Deploy = () => (
  <LiveBuild
    appliedAt={appliedAt}
    build={build}
    buildHref="#build"
    deployHref="#deploy"
  />
)

export const Sync = () => (
  <LiveBuild
    appliedAt={appliedAt}
    build={build}
    buildHref="#build"
    deployHref="#sync"
    variant="sync"
  />
)

export const TagOnly = () => (
  <LiveBuild
    appliedAt={appliedAt}
    build={{ id: 'bld-tag', resolved_tag: '1.14.2' }}
    deployHref="#deploy"
  />
)

export const NoBuild = () => <LiveBuild deployHref="#deploy" />
