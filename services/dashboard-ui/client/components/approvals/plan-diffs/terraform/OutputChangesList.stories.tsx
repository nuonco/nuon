export default {
  title: 'Approvals/PlanDiffs/OutputChangesList',
}

import { OutputChangesList } from './OutputChangesList'

const mockChanges = [
  {
    output: 'database_endpoint',
    action: 'update',
    before: { value: 'db-old.example.com:5432' },
    after: { value: 'db-new.example.com:5432' },
    beforeSensitive: false,
    afterSensitive: false,
    afterUnknown: false,
  },
  {
    output: 'api_secret_key',
    action: 'create',
    before: null,
    after: { value: 'supersecretvalue' },
    beforeSensitive: false,
    afterSensitive: true,
    afterUnknown: false,
  },
  {
    output: 'deprecated_output',
    action: 'delete',
    before: { value: 'old-value' },
    after: null,
    beforeSensitive: false,
    afterSensitive: false,
    afterUnknown: false,
  },
] as any[]

export const Default = () => <OutputChangesList changes={mockChanges} />

export const Empty = () => <OutputChangesList changes={[]} />
