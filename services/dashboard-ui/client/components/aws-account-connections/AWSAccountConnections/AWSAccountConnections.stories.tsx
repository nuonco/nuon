export default { title: 'AWS account connections/AWSAccountConnections' }

import { AWSAccountConnections } from './AWSAccountConnections'

export const Default = () => (
  <AWSAccountConnections
    onSelect={() => {}}
    onVerify={() => {}}
    connections={[
      {
        id: 'awc123',
        created_at: '2026-07-16T00:00:00Z',
        updated_at: '2026-07-16T00:00:00Z',
        name: 'Demo smoke tests',
        account_id: '123456789012',
        default_region: 'us-west-2',
        verification_status: 'verified',
        last_checked_at: '2026-07-16T00:00:00Z',
        role_arn: 'arn:aws:iam::123456789012:role/nuon-smoke-tests',
      },
    ]}
  />
)

export const Empty = () => (
  <AWSAccountConnections
    connections={[]}
    onSelect={() => {}}
    onVerify={() => {}}
  />
)
