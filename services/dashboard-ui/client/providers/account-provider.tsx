import { createContext } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAccount } from '@/lib'
import type { TAccount, TUserJourney } from '@/types'

interface AccountContextValue {
  account: TAccount | null
  isLoading: boolean
  error: any
  refreshAccount: () => void
}

export const AccountContext = createContext<AccountContextValue | undefined>(
  undefined
)

export function AccountProvider({
  children,
  shouldPoll = false,
}: {
  children: React.ReactNode
  shouldPoll?: boolean
}) {
  const { data: account, error, isLoading, refetch } = useQuery({
    queryKey: ['account'],
    queryFn: getAccount,
    refetchInterval: shouldPoll
      ? (query) => {
          const data = query.state.data as TAccount | undefined
          if (!data) return 5000
          const evaluationJourney = (data as any)?.user_journeys?.find(
            (j: TUserJourney) => j.name === 'evaluation'
          )
          if (!evaluationJourney) return 20000
          const hasIncompleteSteps = evaluationJourney.steps.some(
            (s: any) => !s.complete
          )
          return hasIncompleteSteps ? 5000 : 20000
        }
      : false,
  })

  return (
    <AccountContext.Provider
      value={{
        account: account ?? null,
        isLoading,
        error,
        refreshAccount: refetch,
      }}
    >
      {children}
    </AccountContext.Provider>
  )
}
