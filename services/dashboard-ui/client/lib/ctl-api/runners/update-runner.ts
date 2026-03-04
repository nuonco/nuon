import { api } from '@/lib/api'
import type { TRunnerJob } from '@/types'

export interface IUpdateRunnerBody {
  container_image_tag: string
  container_image_url: string
  org_awsiam_role_arn: string
  org_k8s_service_account_name: string
  runner_api_url: string
}

export async function updateRunner({
  body,
  orgId,
  runnerId,
}: {
  body: IUpdateRunnerBody
  runnerId: string
  orgId: string
}) {
  return api<TRunnerJob>({
    body,
    method: 'PATCH',
    orgId,
    path: `runners/${runnerId}/settings`,
  })
}
