'use server'

import { cookies } from 'next/headers'

export async function setPageSidebarCookie(isOpen: boolean) {
  const cookieStore = await cookies()
  cookieStore.set('page_sidebar_open', isOpen ? '1' : '0', {
    path: '/',
    httpOnly: false,
    maxAge: 60 * 60 * 24 * 365,
    sameSite: 'lax',
  })
}

export async function getIsPageSidebarOpenFromCookie(): Promise<boolean> {
  const cookieStore = await cookies()
  return cookieStore.get('page_sidebar_open')?.value === '1'
}
