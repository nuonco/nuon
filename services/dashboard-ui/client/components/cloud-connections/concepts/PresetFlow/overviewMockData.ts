export type TCloudConnectionOverview = {
  id: string
  name: string
  targetId: string
  region: string
  roleArn: string
  preset: 'Stacks' | 'Custom'
  status: 'verified' | 'pending' | 'error'
  statusMessage?: string
  lastVerifiedAt?: string
  createdAt: string
  installCount: number
}

export const CLOUD_CONNECTIONS: TCloudConnectionOverview[] = [
  {
    id: 'cc_01JEXAMPLEPROD',
    name: 'Production stacks',
    targetId: '123456789012',
    region: 'us-east-1',
    roleArn: 'arn:aws:iam::123456789012:role/nuon-production-stacks',
    preset: 'Stacks',
    status: 'verified',
    lastVerifiedAt: '2026-09-28T07:25:00Z',
    createdAt: '2026-08-12T14:10:00Z',
    installCount: 14,
  },
  {
    id: 'cc_01JEXAMPLESTAGE',
    name: 'Staging stacks',
    targetId: '210987654321',
    region: 'us-west-2',
    roleArn: 'arn:aws:iam::210987654321:role/nuon-staging-stacks',
    preset: 'Stacks',
    status: 'pending',
    statusMessage: 'Verification has not run yet.',
    createdAt: '2026-09-28T06:40:00Z',
    installCount: 0,
  },
  {
    id: 'cc_01JEXAMPLECUSTOM',
    name: 'Security-managed role',
    targetId: '555555555555',
    region: 'eu-west-1',
    roleArn:
      'arn:aws:iam::555555555555:role/security/platform/nuon-custom-access',
    preset: 'Custom',
    status: 'verified',
    lastVerifiedAt: '2026-09-27T15:10:00Z',
    createdAt: '2026-09-15T11:05:00Z',
    installCount: 0,
  },
  {
    id: 'cc_01JEXAMPLEERROR',
    name: 'Recovery stacks',
    targetId: '444444444444',
    region: 'us-east-2',
    roleArn: 'arn:aws:iam::444444444444:role/nuon-recovery-stacks',
    preset: 'Stacks',
    status: 'error',
    statusMessage:
      'AWS denied sts:AssumeRoleWithWebIdentity for this connection subject.',
    lastVerifiedAt: '2026-09-28T05:10:00Z',
    createdAt: '2026-09-01T09:30:00Z',
    installCount: 2,
  },
]
