import {
  ConnectionSetupPrototype,
  prototypeConnections,
} from './__stories__/ConnectionSetupPrototype'

export default { title: 'Features / Cloud connections / Setup prototype' }

export const CreateECR = () => (
  <ConnectionSetupPrototype connection={prototypeConnections[0]} />
)
export const CreateGAR = () => (
  <ConnectionSetupPrototype connection={prototypeConnections[1]} />
)
export const CreateACR = () => (
  <ConnectionSetupPrototype connection={prototypeConnections[2]} />
)
export const ExistingConnection = () => (
  <ConnectionSetupPrototype connection={prototypeConnections[0]} existing />
)
export const ECRPolicyPreview = () => (
  <ConnectionSetupPrototype
    connection={prototypeConnections[0]}
    existing
    preview
  />
)
export const GARPolicyPreview = () => (
  <ConnectionSetupPrototype
    connection={prototypeConnections[1]}
    existing
    preview
  />
)
export const ACRPolicyPreview = () => (
  <ConnectionSetupPrototype
    connection={prototypeConnections[2]}
    existing
    preview
  />
)

export const LoadingConnection = () => (
  <ConnectionSetupPrototype
    connection={prototypeConnections[0]}
    existing
    state="loading"
  />
)

export const GeneratingCommands = () => (
  <ConnectionSetupPrototype
    connection={prototypeConnections[0]}
    existing
    preview
    state="generating"
  />
)

export const CommandGenerationError = () => (
  <ConnectionSetupPrototype
    connection={prototypeConnections[0]}
    existing
    preview
    state="error"
  />
)
