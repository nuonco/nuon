'use server'

import { cookies } from 'next/headers'

export async function setPageNavCookie(isOpen: boolean) {
  const cookieStore = await cookies()
  cookieStore.set('is-page-nav-open', Boolean(isOpen).toString())
}
