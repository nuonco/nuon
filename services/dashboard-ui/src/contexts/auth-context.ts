import { createContext } from 'react'
import type { IUser } from '@/types/dashboard.types'

export interface IAuthContext {
  user: IUser | null | undefined
  error?: Error
  isLoading: boolean
  isAdmin: boolean
  useAuthService: boolean
  authServiceUrl?: string
}

export const AuthContext = createContext<IAuthContext | undefined>(undefined)
