'use server'

import { joinWaitlist, type IJoinWaitlist } from '@/lib'
import type { TWaitlist } from '@/types'

export async function requestWaitlistAccess(
  formData: FormData
): Promise<TWaitlist> {
  const data = Object.fromEntries(formData) as IJoinWaitlist

  return joinWaitlist(data)
}
