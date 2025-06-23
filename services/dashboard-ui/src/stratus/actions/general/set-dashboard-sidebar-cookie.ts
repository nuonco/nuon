'use server'

import { cookies } from 'next/headers'

export async function setDashboardSidebarCookie(isOpen: boolean) {
  const cookieStore = await cookies()
  cookieStore.set('is-sidebar-open', Boolean(isOpen).toString())
}
