export const VIEW_ORG_ID = 'org-1'
export const VIEW_APP_ID = 'app-1'
export const VIEW_INSTALL_ID = 'inst-1'

export const viewPath = (suffix = '') =>
  `/${VIEW_ORG_ID}/apps/${VIEW_APP_ID}/installs/${VIEW_INSTALL_ID}${suffix}`

export type TFixtureReply = {
  body: unknown
  paginated?: boolean | { hasNext: boolean; offset: number; limit: number }
  status?: number
  pending?: boolean
  delayMs?: number
}

export type TFixture = (
  url: URL,
  init: RequestInit | undefined
) => TFixtureReply | undefined

const misses: string[] = []

export const fixtureMisses = () => misses.slice()

const json = (
  body: unknown,
  paginated: TFixtureReply['paginated'] = false,
  status = 200
) => {
  const headers = new Headers({ 'content-type': 'application/json' })
  if (paginated) {
    const pagination =
      typeof paginated === 'object'
        ? paginated
        : { hasNext: false, offset: 0, limit: 20 }
    headers.set('X-Nuon-Page-Next', String(pagination.hasNext))
    headers.set('X-Nuon-Page-Offset', String(pagination.offset))
    headers.set('X-Nuon-Page-Limit', String(pagination.limit))
  }
  return new Response(JSON.stringify(body), { status, headers })
}

class StoryEventSource {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSED = 2
  readyState = StoryEventSource.OPEN
  onopen: ((event: Event) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  url: string

  constructor(url: string) {
    this.url = url
    queueMicrotask(() => {
      this.onopen?.(new Event('open'))
    })
  }

  addEventListener() {}
  removeEventListener() {}
  close() {
    this.readyState = StoryEventSource.CLOSED
  }
  dispatchEvent() {
    return true
  }
}

let restoreFetch: typeof window.fetch | null = null
let restoreEventSource: typeof EventSource | null = null
let current: TFixture | null = null
let generation = 0

export const beginInstallFixture = (fixture: TFixture) => {
  misses.length = 0
  current = fixture
  const mine = ++generation
  if (restoreFetch) return mine
  restoreFetch = window.fetch.bind(window)
  window.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
    const raw =
      typeof input === 'string'
        ? input
        : input instanceof URL
          ? input.href
          : input.url
    const url = new URL(raw, window.location.origin)
    if (url.pathname === '/readyz') return json({ status: 'ok' })
    if (url.pathname === '/version') {
      return json({
        api: { version: '0.0.0', git_ref: 'dev' },
        ui: { version: '0.0.0' },
      })
    }
    const isApi = url.pathname.startsWith('/v1/')
    if ((isApi || url.pathname.startsWith('/api/')) && current) {
      const reply = current(url, init)
      if (reply?.pending) return new Promise(() => {})
      if (reply?.delayMs) {
        await new Promise((resolve) => setTimeout(resolve, reply.delayMs))
      }
      if (reply) return json(reply.body, reply.paginated, reply.status ?? 200)
      if (isApi) {
        misses.push(`${init?.method ?? 'GET'} ${url.pathname}${url.search}`)
        return json(
          { error: 'fixture miss', description: url.pathname },
          false,
          404
        )
      }
    }
    return restoreFetch!(input, init)
  }) as typeof window.fetch
  restoreEventSource = window.EventSource
  window.EventSource = StoryEventSource as unknown as typeof EventSource
  return mine
}

export const endInstallFixture = (mine: number) => {
  if (mine !== generation) return
  current = null
  if (restoreFetch) {
    window.fetch = restoreFetch
    restoreFetch = null
  }
  if (restoreEventSource) {
    window.EventSource = restoreEventSource
    restoreEventSource = null
  }
}
