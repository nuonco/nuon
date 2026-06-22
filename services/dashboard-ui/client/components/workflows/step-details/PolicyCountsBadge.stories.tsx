export default {
  title: 'Workflows/PolicyCountsBadge',
}

import { PolicyCountsBadge } from './PolicyCountsBadge'
import type { TWorkflowStep } from '@/types'

const warningStep = {
  id: 'step-warn',
  status: {
    status: 'success',
    metadata: {
      warn_violations: [
        {
          policy_id: 'pol-public-endpoint',
          message:
            "EKS cluster 'module.eks.aws_eks_cluster.this[0]' has public endpoint access enabled - ensure this is intentional for demo/development environments",
          severity: 'warn',
        },
      ],
      passed_policy_ids: ['pol-tags', 'pol-region'],
    },
  },
} as TWorkflowStep

const violationStep = {
  id: 'step-deny',
  status: {
    status: 'error',
    metadata: {
      deny_violations: [
        {
          policy_id: 'pol-public-read',
          message:
            "S3 bucket 'module.storage.aws_s3_bucket.artifacts[0]' must not allow public read access - set 'block_public_acls' and 'restrict_public_buckets' to true before this can be applied",
          severity: 'deny',
        },
        {
          policy_id: 'pol-encryption',
          message:
            "EBS volume 'module.eks.aws_ebs_volume.data' is not encrypted at rest - all persistent volumes must use a customer-managed KMS key per the security baseline",
          severity: 'deny',
        },
      ],
      warn_violations: [
        {
          policy_id: 'pol-public-endpoint',
          message:
            "EKS cluster 'module.eks.aws_eks_cluster.this[0]' has public endpoint access enabled - ensure this is intentional for demo/development environments",
          severity: 'warn',
        },
      ],
    },
  },
} as TWorkflowStep

const passedStep = {
  id: 'step-passed',
  status: {
    status: 'success',
    metadata: {
      passed_policy_ids: ['pol-tags', 'pol-region', 'pol-cost'],
    },
  },
} as TWorkflowStep

const noPolicyStep = {
  id: 'step-none',
} as TWorkflowStep

const Center = ({ children }: { children: React.ReactNode }) => (
  <div className="flex min-h-[60vh] items-center justify-center">
    {children}
  </div>
)

export const SingleWarning = () => (
  <Center>
    <PolicyCountsBadge step={warningStep} />
  </Center>
)

export const DenyAndWarn = () => (
  <Center>
    <PolicyCountsBadge step={violationStep} />
  </Center>
)

export const Passed = () => (
  <Center>
    <PolicyCountsBadge step={passedStep} />
  </Center>
)

export const NoPolicy = () => (
  <Center>
    <PolicyCountsBadge step={noPolicyStep} />
  </Center>
)
