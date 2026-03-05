const MAX_AGE = 60 * 60 * 24 * 365

function setCookie(name: string, value: string) {
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; SameSite=Lax; max-age=${MAX_AGE}`
}

function getCookie(name: string): string | undefined {
  const match = document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith(`${name}=`))
  return match ? decodeURIComponent(match.slice(name.length + 1)) : undefined
}

export const getOrgSession = () => getCookie('org_session')
export const setOrgSession = (orgId: string) => setCookie('org_session', orgId)

export const getSidebarOpen = () => getCookie('sidebar_open') === '1'
export const setSidebarOpen = (isOpen: boolean) =>
  setCookie('sidebar_open', isOpen ? '1' : '0')

export const getPageSidebarOpen = () => getCookie('page_sidebar_open') === '1'
export const setPageSidebarOpen = (isOpen: boolean) =>
  setCookie('page_sidebar_open', isOpen ? '1' : '0')
