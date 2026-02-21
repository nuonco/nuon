import '@/app/old-styles.css'
import '@/app/globals.css'
import React, { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { AUTH_SERVICE_URL, APP_URL } from '@/configs/auth'
import { apiClient } from '@/lib/api-client'
import { AccountProvider } from '@/providers/account-provider'
import { UserJourneyContext } from '@/providers/user-journey-provider'
import { AuthContext } from '@/contexts/auth-context'
import { AppRouter } from '@/routes/index'
import type { IUser, TAccount } from '@/types'

function AppBootstrap() {
  const [initialUser, setInitialUser] = useState<IUser | null>(null)
  const [initialAccount, setInitialAccount] = useState<TAccount | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function bootstrap() {
      try {
        const { data: account, error: accountError, status } = await apiClient<TAccount>({
          path: '/api/account',
        })

        if (status === 401 || accountError?.error === 'unauthorized') {
          window.location.href = `${AUTH_SERVICE_URL}/?url=${APP_URL}`
          return
        }

        if (accountError || !account) {
          setError('Failed to load account')
          return
        }

        const user: IUser = {
          sub: account.id,
          email: account.email,
          name: account.name || account.email,
          picture: undefined, // Identity picture not available via ctl-api; Avatar uses initials
        }

        setInitialUser(user)
        setInitialAccount(account)
        setIsLoading(false)
      } catch (err) {
        console.error('Bootstrap error:', err)
        setError(err instanceof Error ? err.message : 'Unknown error')
        setIsLoading(false)
      }
    }

    bootstrap()
  }, [])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900 mx-auto mb-4" />
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <p className="text-red-600 mb-4">Error: {error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  const isAdmin = initialUser?.email?.endsWith('@nuon.co') ?? false

  return (
    <AuthContext.Provider
      value={{
        user: initialUser,
        error: undefined,
        isLoading: false,
        isAdmin,
        useAuthService: true,
        authServiceUrl: AUTH_SERVICE_URL,
      }}
    >
      <AccountProvider
        initAccount={initialAccount}
        shouldPoll={true}
        useDynamicPolling={true}
      >
        <UserJourneyContext.Provider value={{ isBYOC: false, isCustomerPortalEnabled: false }}>
          <AppRouter />
        </UserJourneyContext.Provider>
      </AccountProvider>
    </AuthContext.Provider>
  )
}

const container = document.getElementById('root')
if (container) {
  const root = createRoot(container)
  root.render(
    <React.StrictMode>
      <AppBootstrap />
    </React.StrictMode>
  )
}
