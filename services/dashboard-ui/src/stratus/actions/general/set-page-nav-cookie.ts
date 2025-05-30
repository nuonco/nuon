'use server'

import { cookies } from 'next/headers'

export async function setPageNavCookie(isOpen: boolean) {
  cookies().set('is-page-nav-open', Boolean(isOpen).toString())
}
