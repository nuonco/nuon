export default {
  title: 'Installs/Forms/PlatformFields',
}

import type { TAWSAccountConnection } from '@/types'
import { PlatformFields } from './PlatformFields'

const connections = [
  {
    id: 'awsconn000000000000000001',
    name: 'prod-account',
    account_id: '123456789012',
    verification_status: 'verified',
  },
  {
    id: 'awsconn000000000000000002',
    name: 'staging-account',
    account_id: '210987654321',
    verification_status: 'pending',
  },
] as TAWSAccountConnection[]

export const AWS = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="aws" />
  </form>
)

export const Azure = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="azure" />
  </form>
)

export const GCP = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="gcp" />
  </form>
)

export const AWSwithDraft = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="aws" draftValues={{ region: 'us-west-2' }} />
  </form>
)

export const AWSTargetAccountRequired = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="aws" requireTargetAccount />
  </form>
)

export const AWSTargetAccountFromConnection = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields
      platform="aws"
      requireTargetAccount
      awsAccountConnections={connections}
      draftValues={{ aws_connection_id: 'awsconn000000000000000001' }}
    />
  </form>
)

export const AzureTargetSubscriptionRequired = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="azure" requireTargetAccount />
  </form>
)

export const GCPTargetProjectRequired = () => (
  <form className="max-w-2xl p-6">
    <PlatformFields platform="gcp" requireTargetAccount />
  </form>
)
