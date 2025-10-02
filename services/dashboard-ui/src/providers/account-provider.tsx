'use client'

import { createContext, useState } from 'react'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
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
}: {
  children: React.ReactNode
  initAccount?: TAccount
} & IPollingProps) {
  const [refresh, shouldRefresh] = useState<number>(0)
  const {
    data: account,
    error,
    isLoading,
  } = usePolling<TAccount>({
    dependencies: [initAccount, refresh],
    initData: initAccount,
    path: `/api/account`,
    pollInterval,
    shouldPoll,
  })

  const refreshAccount = async () => {
    shouldRefresh((prev) => prev + 1)
  }

  return (
    <AccountContext.Provider
      value={{ account, isLoading, error, refreshAccount }}
    >
      {children}
    </AccountContext.Provider>
  )
}
