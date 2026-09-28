import type { ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import type { TCloudConnection } from '@/types'

export const connection: TCloudConnection = {
  id: 'cc-example',
  org_id: 'org-mock-001',
  name: 'acme-production',
  platform: 'aws',
  target_id: '123456789012',
  principal: 'arn:aws:iam::123456789012:role/acme-nuon',
  created_at: '2026-09-28T07:00:00Z',
  updated_at: '2026-09-28T07:00:00Z',
  status: 'pending',
  verification_in_progress: false,
  preset: 'stacks',
  used_by: { installs: 0 },
  setup: {
    issuer_url: 'https://api.example.com',
    subject: 'org:org-mock-001:connection:cc-example',
    audience: 'sts.amazonaws.com',
    preset: 'stacks',
    trust_policy: {
      Version: '2012-10-17',
      Statement: [
        {
          Effect: 'Allow',
          Action: 'sts:AssumeRoleWithWebIdentity',
          Principal: {
            Federated:
              'arn:aws:iam::123456789012:oidc-provider/api.example.com',
          },
          Condition: {
            StringEquals: {
              'api.example.com:aud': 'sts.amazonaws.com',
              'api.example.com:sub': 'org:org-mock-001:connection:cc-example',
            },
          },
        },
      ],
    },
    permissions_policy: {
      Version: '2012-10-17',
      Statement: [
        {
          Effect: 'Allow',
          Action: ['cloudformation:DescribeStacks'],
          Resource: '*',
        },
      ],
    },
    cli: 'aws iam create-role --role-name <role-name> --assume-role-policy-document file://trust.json',
    terraform:
      'resource "aws_iam_role" "nuon" {\n  name = "<role-name>"\n  assume_role_policy = file("trust.json")\n}',
    cloudformation:
      'NuonRole:\n  Type: AWS::IAM::Role\n  Properties:\n    RoleName: <role-name>',
  },
}
export const customConnection: TCloudConnection = {
  ...connection,
  name: 'acme-custom',
  preset: 'custom',
  setup: {
    ...connection.setup,
    preset: 'custom',
    permissions_policy: undefined,
  },
}
export const failedConnection: TCloudConnection = {
  ...connection,
  status: 'error',
  status_message:
    'AWS denied sts:AssumeRoleWithWebIdentity for this connection subject.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const verifiedConnection: TCloudConnection = {
  ...connection,
  status: 'verified',
  status_message: 'OIDC exchange and read-only CloudFormation probe passed.',
  last_verified_at: '2026-09-28T07:10:00Z',
}
export const noop = () => {}

export const PageStory = ({ children }: { children: ReactNode }) => (
  <NotificationContext.Provider
    value={{
      emitNotification: async () => false,
      permission: 'default',
      requestPermission: async () => 'default',
      isSupported: false,
      settings: { permissionRequested: false },
      hasRequestedPermission: false,
      muted: false,
      toggleMute: noop,
    }}
  >
    <BreadcrumbContext.Provider
      value={{ breadcrumbLinks: [], isLoading: false, updateBreadcrumb: noop }}
    >
      <SidebarContext.Provider
        value={{
          isSidebarOpen: true,
          closeSidebar: noop,
          openSidebar: noop,
          toggleSidebar: noop,
        }}
      >
        {children}
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)
