'use server'

import { cookies } from 'next/headers'

export async function setSidebarCookie(isOpen: boolean) {
  const cookieStore = await cookies()
  cookieStore.set('sidebar_open', isOpen ? '1' : '0', {
    path: '/',
    httpOnly: false,
    maxAge: 60 * 60 * 24 * 365,
    sameSite: 'lax',
  })
}

export async function getIsSidebarOpenFromCookie(): Promise<boolean> {
  const cookieStore = await cookies()
  return cookieStore.get('sidebar_open')?.value === '1'
}
