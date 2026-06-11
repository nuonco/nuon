import { api } from '@/lib/api'
import type { TWorkflowStep } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

const PAGE_LIMIT = 100

export const getWorkflowSteps = async ({
  workflowId,
  orgId,
}: {
  workflowId: string
  orgId: string
}) => {
  const steps: TWorkflowStep[] = []
  let offset = 0

  for (;;) {
    const page = await api<TWorkflowStep[]>({
      path: `workflows/${workflowId}/steps${buildQueryParams({ limit: PAGE_LIMIT, offset })}`,
      orgId,
    })

    if (!page?.length) break
    steps.push(...page)
    if (page.length < PAGE_LIMIT) break
    offset += PAGE_LIMIT
  }

  return steps
}
