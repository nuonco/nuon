import { useState } from 'react'
import { ConnectionSetupMode } from './ConnectionSetupMode'

export default { title: 'Features / Cloud connections / Setup mode' }

export const AWS = () => {
  const [existing, setExisting] = useState(false)
  return (
    <ConnectionSetupMode
      platform="aws"
      existingPrincipal={existing}
      onChange={setExisting}
    />
  )
}

export const ExistingAWS = () => {
  const [existing, setExisting] = useState(true)
  return (
    <ConnectionSetupMode
      platform="aws"
      existingPrincipal={existing}
      onChange={setExisting}
    />
  )
}

export const GCP = () => {
  const [existing, setExisting] = useState(true)
  return (
    <ConnectionSetupMode
      platform="gcp"
      existingPrincipal={existing}
      onChange={setExisting}
    />
  )
}
