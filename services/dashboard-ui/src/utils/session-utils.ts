import type { IUser } from '@/types/dashboard.types'

export const isNuonSession = (user: IUser): boolean => {
  return user?.email ? user?.email?.endsWith('@nuon.co') : false
}
