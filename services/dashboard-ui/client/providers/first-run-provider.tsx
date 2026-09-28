import { createContext, type ReactNode } from 'react'
import type { TCloud, TPath } from '@/components/onboarding/first-run/constants'
import type { IFirstRunJourney } from '@/hooks/use-first-run-journey'

export interface IFirstRunContext {
  orgId: string
  journey: Pick<IFirstRunJourney, 'metadata' | 'saveStep' | 'complete'>
  // Set when the GitHub App callback returned to onboarding.
  vcsConnectionId?: string
  vcsError?: boolean
  // The stepper follows the path, so choosing one swaps the wizard's steps.
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
