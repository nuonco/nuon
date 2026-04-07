export default {
  title: 'Approvals/PlanDiffs/ResourceChangesList',
}

import { ResourceChangesList } from './ResourceChangesList'

const mockChanges = [
  {
    address: 'aws_s3_bucket.app_assets',
    resource: 'aws_s3_bucket',
    name: 'app_assets',
    action: 'create',
    module: null,
    before: null,
    after: { bucket: 'my-app-assets', acl: 'private' },
  },
  {
    address: 'aws_instance.web_server',
    resource: 'aws_instance',
    name: 'web_server',
    action: 'update',
    module: 'module.networking',
    before: { instance_type: 't3.micro' },
    after: { instance_type: 't3.small' },
  },
  {
    address: 'aws_db_instance.legacy',
    resource: 'aws_db_instance',
    name: 'legacy',
    action: 'delete',
    module: null,
    before: { allocated_storage: 20, engine: 'postgres' },
    after: null,
  },
  {
    address: 'aws_iam_role.service_role',
    resource: 'aws_iam_role',
    name: 'service_role',
    action: 'read',
    module: null,
    before: null,
    after: null,
  },
] as any[]

export const Default = () => <ResourceChangesList changes={mockChanges} />

export const Empty = () => <ResourceChangesList changes={[]} />
