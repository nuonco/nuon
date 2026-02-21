import { setCookie, getCookie } from '@/utils/cookies'

class CookieStore {
  get(name: string) {
    const value = getCookie(name)
    return value ? { name, value } : undefined
  }

  set(name: string, value: string, options?: { path?: string; maxAge?: number }) {
    const days = options?.maxAge ? options.maxAge / 86400 : 365
    setCookie(name, value, days)
  }

  delete(name: string) {
    setCookie(name, '', -1)
  }
}

export async function cookies() {
  return new CookieStore()
}

export async function headers() {
  return new Headers()
}
