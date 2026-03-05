import { api } from '@/lib/api'
import type { TRunnerSettings } from '@/types'

export const getRunnerSettings = ({
  runnerId,
  orgId,
}: {
  runnerId: string
  orgId: string
}) =>
  api<TRunnerSettings>({
    path: `runners/${runnerId}/settings`,
    orgId,
  })
