import { api } from '@/lib/api'

export type TPruneRunnerTokensResponse = {
  invalidated_count: number
}

export async function pruneRunnerTokens({
  runnerId,
  orgId,
}: {
  runnerId: string
  orgId: string
}) {
  return api<TPruneRunnerTokensResponse>({
    method: 'POST',
    orgId,
    path: `runners/${runnerId}/prune-tokens`,
  })
}
