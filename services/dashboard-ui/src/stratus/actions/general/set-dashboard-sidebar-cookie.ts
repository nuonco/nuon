'use server'

import { cookies } from 'next/headers'

export async function setDashboardSidebarCookie(isOpen: boolean) {
  cookies().set('is-sidebar-open', Boolean(isOpen).toString())
}
