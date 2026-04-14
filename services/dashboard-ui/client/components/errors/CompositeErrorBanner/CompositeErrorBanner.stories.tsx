import { CompositeErrorBanner } from './CompositeErrorBanner'
import type { TCompositeError } from '@/types'

export default { title: 'Errors/CompositeErrorBanner' }

const terraformError: TCompositeError = {
  owner_type: 'apply',
  severity: 'critical',
  summary: 'Error creating S3 bucket: BucketAlreadyExists',
  detail: `Error: creating Amazon S3 (Simple Storage) Bucket (my-app-bucket): operation error S3: CreateBucket, https response error StatusCode: 409, RequestID: EXAMPLE123, HostID: abc123, BucketAlreadyOwnedByYou: Your previous request to create the named bucket succeeded and you already own it.

  with aws_s3_bucket.main,
  on main.tf line 12, in resource "aws_s3_bucket" "main":
  12: resource "aws_s3_bucket" "main" {`,
}

const helmError: TCompositeError = {
  owner_type: 'apply',
  severity: 'critical',
  summary: 'Helm release failed: timed out waiting for release to complete',
  detail: `Error: timed out waiting for the condition

  LAST DEPLOYED: Mon Jan 13 12:34:56 2025
  NAMESPACE: my-app
  STATUS: failed
  REVISION: 3

  NOTES:
  Pod my-app-deployment-6d8f4b9c7-xk2p9 is in CrashLoopBackOff state.`,
}

const runnerError: TCompositeError = {
  owner_type: 'runner',
  severity: 'warning',
  summary: 'Runner lost connection during deployment',
}

export const NoErrors = () => <CompositeErrorBanner errors={[]} />

export const NullErrors = () => <CompositeErrorBanner errors={null} />

export const SingleError = () => <CompositeErrorBanner errors={[terraformError]} />

export const SingleErrorNoDetail = () => <CompositeErrorBanner errors={[runnerError]} />

export const MultipleErrors = () => (
  <CompositeErrorBanner errors={[terraformError, helmError, runnerError]} />
)
