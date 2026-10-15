export default {
  title: 'Playground / Fleet',
  fullBleed: true,
}

import { InstallFleetPlayground } from './InstallFleetPlayground'
import { mixedFleet } from './fixtures'

const pageStyle = {
  height: '100vh',
  display: 'flex',
  flexDirection: 'column' as const,
  overflow: 'auto',
}

export const MixedFleet = () => (
  <div style={pageStyle}>
    <InstallFleetPlayground installs={mixedFleet} currentBranch="main" />
  </div>
)
MixedFleet.storyName = 'Mixed fleet'

export const Empty = () => (
  <div style={pageStyle}>
    <InstallFleetPlayground installs={[]} />
  </div>
)
