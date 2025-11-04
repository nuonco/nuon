import type { User } from '@auth0/nextjs-auth0/types'

export const isNuonSession = (user: User): boolean => {
  return user?.email ? user?.email?.endsWith('@nuon.co') : false
}
