export default {
  title: 'Playground / Installs / Deployment details v2',
}

import { DeploymentDetailV2 } from './DeploymentDetailV2'
import { deploymentScenario } from './fixtures'
import { SimulatedDeployment } from './SimulatedRun'

export const Loading = () => <DeploymentDetailV2 isLoading />

export const Active = () => <SimulatedDeployment />

export const Finished = () => (
  <DeploymentDetailV2 {...deploymentScenario('finished')} />
)

export const Failed = () => (
  <DeploymentDetailV2 {...deploymentScenario('failed')} />
)
