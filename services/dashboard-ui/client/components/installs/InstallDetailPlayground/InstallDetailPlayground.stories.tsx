export default {
  title: 'Installs/InstallDetailPlayground',
  fullBleed: true,
}

import { InstallDetailPlayground } from './InstallDetailPlayground'
import {
  configCurrentFixture,
  branchMovedFixture,
  resourceLagFixture,
  infraDriftFixture,
} from './fixtures'

const pageStyle = {
  height: '100vh',
  display: 'flex',
  flexDirection: 'column' as const,
  overflow: 'hidden',
}

/**
 * Config-file-managed install — everything in sync. Branch, config, resources,
 * and infrastructure all match expected state. Drift: none.
 */
export const ConfigCurrent = () => (
  <div style={pageStyle}>
    <InstallDetailPlayground
      install={configCurrentFixture}
      className="flex-1 rounded-none border-0"
    />
  </div>
)

ConfigCurrent.storyName = 'Config current'

/**
 * Expected branch has moved to feat/multi-region. Applied is still main.
 * Stack, sandbox, api, and worker are all pending the new branch target.
 * Frontend is unchanged. Drift: none.
 */
export const BranchMoved = () => (
  <div style={pageStyle}>
    <InstallDetailPlayground
      install={branchMovedFixture}
      className="flex-1 rounded-none border-0"
    />
  </div>
)

BranchMoved.storyName = 'Branch moved'

/**
 * Branch is current (main) but api and worker components (and their images)
 * are mid-deploy and haven't applied the latest patch. Config lag: api, worker.
 * Sandbox and stack are current. Drift: none.
 */
export const ResourceLag = () => (
  <div style={pageStyle}>
    <InstallDetailPlayground
      install={resourceLagFixture}
      className="flex-1 rounded-none border-0"
    />
  </div>
)

ResourceLag.storyName = 'Resource lag'

/**
 * Config is fully current — branch, stack, sandbox, components, and images all
 * match expected. Infrastructure drift detected on sandbox and the cache component.
 * Config lag: none.
 */
export const InfraDrift = () => (
  <div style={pageStyle}>
    <InstallDetailPlayground
      install={infraDriftFixture}
      className="flex-1 rounded-none border-0"
    />
  </div>
)

InfraDrift.storyName = 'Infra drift'
