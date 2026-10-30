import { AddRepositoryAccess } from './AddRepositoryAccess'
import { RepositoryAccessDemo } from '../__stories__/RepositoryAccessDemo'
import {
  acrConnection,
  ecrConnection,
  garConnection,
  noop,
} from '../__stories__/fixtures'

export default { title: 'Features / Cloud connections / Repository access' }

export const ECR = () => <RepositoryAccessDemo connection={ecrConnection} />
export const GAR = () => <RepositoryAccessDemo connection={garConnection} />
export const ACR = () => <RepositoryAccessDemo connection={acrConnection} />
export const ECRGenerated = () => (
  <RepositoryAccessDemo connection={ecrConnection} generated />
)
export const GARGenerated = () => (
  <RepositoryAccessDemo connection={garConnection} generated />
)
export const ACRGenerated = () => (
  <RepositoryAccessDemo connection={acrConnection} generated />
)
export const Generating = () => (
  <AddRepositoryAccess
    connection={ecrConnection}
    repositories="acme/api"
    onGenerate={noop}
    isPending
  />
)
export const Failed = () => (
  <AddRepositoryAccess
    connection={ecrConnection}
    repositories="acme/api"
    onGenerate={noop}
    error={{
      error: 'Permission command generation failed',
      description: 'Unable to load setup commands. Try again.',
      user_error: false,
    }}
  />
)
export const InvalidNames = () => (
  <AddRepositoryAccess
    connection={ecrConnection}
    repositories="acme/*:latest"
    onGenerate={noop}
  />
)
export const MissingRegion = () => (
  <AddRepositoryAccess
    connection={{ ...ecrConnection, default_region: undefined }}
    onGenerate={noop}
  />
)
