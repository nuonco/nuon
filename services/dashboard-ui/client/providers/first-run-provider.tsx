import { createContext, type ReactNode } from 'react'
import type { TCloud, TPath } from '@/components/onboarding/first-run/constants'
import type { IFirstRunJourney } from '@/hooks/use-first-run-journey'

export interface IFirstRunContext {
  orgId: string
  journey: Pick<IFirstRunJourney, 'metadata' | 'saveStep' | 'complete'>
  vcsConnectionId?: string
  vcsError?: boolean
  choosePath: (path: TPath, cloud: TCloud) => void
  backToIntro: () => void
}

export const FirstRunContext = createContext<IFirstRunContext | undefined>(undefined)

export function FirstRunProvider({
  value,
  children,
}: {
  value: IFirstRunContext
  children: ReactNode
}) {
  return <FirstRunContext.Provider value={value}>{children}</FirstRunContext.Provider>
}
