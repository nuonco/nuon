import { ConnectionPolicies } from './ConnectionPolicies'
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
export const ECR = () => <ConnectionPolicies connection={ecrConnection} />
export const GAR = () => <ConnectionPolicies connection={garConnection} />
export const ACR = () => <ConnectionPolicies connection={acrConnection} />
