import { ConnectionPolicies } from './ConnectionPolicies'
import { RepositoryAccessDemo } from './__stories__/RepositoryAccessDemo'
import {
  acrConnection,
  connection,
  customConnection,
  ecrConnection,
  garConnection,
} from './__stories__/fixtures'

export default { title: 'Features / Cloud connections / Policies' }

export const Stacks = () => <ConnectionPolicies connection={connection} />
export const Custom = () => <ConnectionPolicies connection={customConnection} />
export const ECR = () => (
  <ConnectionPolicies
    connection={ecrConnection}
    repositoryAccess={<RepositoryAccessDemo connection={ecrConnection} />}
  />
)
export const GAR = () => (
  <ConnectionPolicies
    connection={garConnection}
    repositoryAccess={<RepositoryAccessDemo connection={garConnection} />}
  />
)
export const ACR = () => (
  <ConnectionPolicies
    connection={acrConnection}
    repositoryAccess={<RepositoryAccessDemo connection={acrConnection} />}
  />
)
