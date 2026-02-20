'use client'

import { useMutation } from '@/hooks/use-mutation'
import { useServerAction } from '@/hooks/use-server-action'
import type { TAPIResponse } from '@/types'

const DASHBOARD_MODE = process.env.NEXT_PUBLIC_DASHBOARD_MODE || 'nextjs'

/**
 * useAction — compatibility wrapper that delegates to useServerAction (Next.js mode)
 * or useMutation (Go BFF mode) based on NEXT_PUBLIC_DASHBOARD_MODE.
 *
 * Usage:
 *   const { execute } = useAction({
 *     action: shutdownRunner,                           // Next.js server action
 *     endpoint: '/api/actions/runners/shutdown-runner',  // Go BFF endpoint
 *   })
 *   execute({ runnerId, orgId })
 */
export function useAction<TArgs, TData>({
  action,
  endpoint,
}: {
  action: (...args: any[]) => Promise<TAPIResponse<TData>>
  endpoint: string
}) {
  if (DASHBOARD_MODE === 'go') {
    // eslint-disable-next-line react-hooks/rules-of-hooks
    return useMutation<TArgs, TData>(endpoint)
  }

  // eslint-disable-next-line react-hooks/rules-of-hooks
  return useServerAction<[TArgs], TData>({ action: action as any })
}
