'use client'

import { createContext, useState } from 'react'
import { useQuery, } from '@/hooks/use-query'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useJourneyPollingInterval } from '@/hooks/use-journey-polling-interval'
import type { TAccount } from '@/types'

interface AccountContextValue {
  account: TAccount | null
  isLoading: boolean
  error: any
  refreshAccount: () => Promise<void>
}

export const AccountContext = createContext<AccountContextValue | undefined>(
  undefined
)

export function AccountProvider({
  children,
  initAccount,
  pollInterval = 20000,
  shouldPoll = false,
  useDynamicPolling = true,
}: {
  children: React.ReactNode
  initAccount?: TAccount
  /**
   * Enable dynamic polling based on journey state (default: true)
   * When true, uses 5s during onboarding, 20s when complete
   * When false, uses fixed pollInterval (default: 20s)
   */
  useDynamicPolling?: boolean
} & IPollingProps) {
  const [refresh, shouldRefresh] = useState<number>(0)

  // Get initial account data to determine polling behavior
  const {
    data: account,
    error,
    isLoading,
  } = useQuery<TAccount>({
    dependencies: [initAccount, refresh],
    initData: initAccount,
    path: `/api/account`,
  })

  // Calculate dynamic polling interval based on current account data
  const dynamicInterval = useJourneyPollingInterval(account)

  // Use the dynamic interval if dynamic polling is enabled, otherwise use static interval
  const effectivePollInterval = useDynamicPolling ? dynamicInterval : pollInterval

  // Main polling with dynamic interval (this will restart when interval changes)
  const {
    data: finalAccount,
    error: finalError,
    isLoading: finalIsLoading,
  } = usePolling<TAccount>({
    dependencies: [account, refresh, effectivePollInterval], // Include interval in dependencies
    initData: account || initAccount,
    path: `/api/account`,
    pollInterval: effectivePollInterval,
    shouldPoll: shouldPoll && !!account, // Only start after we have account data
  })

  const refreshAccount = async () => {
    shouldRefresh((prev) => prev + 1)
  }

  // Use final account data, or fall back to initial data if final polling hasn't started yet
  return (
    <AccountContext.Provider
      value={{
        account: finalAccount || account,
        isLoading: finalIsLoading || isLoading,
        error: finalError || error,
        refreshAccount
      }}
    >
      {children}
    </AccountContext.Provider>
  )
}
