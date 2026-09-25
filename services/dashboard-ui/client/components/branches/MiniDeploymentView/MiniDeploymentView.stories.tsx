export default {
  title: 'Branches/MiniDeploymentView',
}

import { MiniDeploymentView } from './MiniDeploymentView'
import {
  miniDeployGroups,
  miniDeployInstalls,
} from './MiniDeploymentView.fixtures'

export const Default = () => (
  <div className="max-w-2xl p-6">
    <MiniDeploymentView
      groups={miniDeployGroups}
      installs={miniDeployInstalls}
      showRollout
    />
  </div>
)
