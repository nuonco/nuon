export default {
  title: 'Features / Installs / Detail playground',
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

export const ConfigCurrent = () => (
  <div style={pageStyle}>
    <InstallDetailPlayground
      install={configCurrentFixture}
      className="flex-1 rounded-none border-0"
    />
  </div>
)

ConfigCurrent.storyName = 'Config current'

export const BranchMoved = () => (
  <div style={pageStyle}>
    <InstallDetailPlayground
      install={branchMovedFixture}
      className="flex-1 rounded-none border-0"
    />
  </div>
)

BranchMoved.storyName = 'Branch moved'

export const ResourceLag = () => (
  <div style={pageStyle}>
    <InstallDetailPlayground
      install={resourceLagFixture}
      className="flex-1 rounded-none border-0"
    />
  </div>
)

ResourceLag.storyName = 'Resource lag'

export const InfraDrift = () => (
  <div style={pageStyle}>
    <InstallDetailPlayground
      install={infraDriftFixture}
      className="flex-1 rounded-none border-0"
    />
  </div>
)

InfraDrift.storyName = 'Infra drift'
