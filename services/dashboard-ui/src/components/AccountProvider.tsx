'use client'

import { createContext, useContext, useEffect, useState } from 'react'
import type { TAccount } from '@/types'
import { getCurrentAccount } from '@/components/accounts-actions'

interface AccountContextType {
  account: TAccount | null
  loading: boolean
  error: string | null
  refreshAccount: () => Promise<void>
}

const AccountContext = createContext<AccountContextType | undefined>(undefined)

export function AccountProvider({ children }: { children: React.ReactNode }) {
  const [account, setAccount] = useState<TAccount | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchAccount = async () => {
    try {
      setError(null)
      const accountData = await getCurrentAccount()
      setAccount(accountData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch account')
      console.error('Error fetching account:', err)
    } finally {
      setLoading(false)
    }
  }

  const refreshAccount = async () => {
    setLoading(true)
    await fetchAccount()
  }

  useEffect(() => {
    fetchAccount()
  }, [])

  // Optional: Add polling for real-time updates
  useEffect(() => {
    const interval = setInterval(() => {
      if (!loading) {
        fetchAccount()
      }
    }, 10000) // Poll every 10 seconds

    return () => clearInterval(interval)
  }, [loading])

  return (
    <AccountContext.Provider
      value={{ account, loading, error, refreshAccount }}
    >
      {children}
    </AccountContext.Provider>
  )
}

export function useAccount() {
  const context = useContext(AccountContext)
  if (context === undefined) {
    throw new Error('useAccount must be used within an AccountProvider')
  }
  return context
}

