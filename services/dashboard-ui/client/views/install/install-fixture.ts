export const VIEW_ORG_ID = 'org-1'
export const VIEW_APP_ID = 'app-1'
export const VIEW_INSTALL_ID = 'inst-1'

export const viewPath = (suffix = '') =>
  `/${VIEW_ORG_ID}/apps/${VIEW_APP_ID}/installs/${VIEW_INSTALL_ID}${suffix}`

export type TFixtureReply = {
  body: unknown
  paginated?: boolean
  status?: number
  pending?: boolean
  events?: Array<{ name: string; data: unknown; delay?: number }>
}

export type TFixture = (
  url: URL,
  init: RequestInit | undefined
) => TFixtureReply | undefined

const misses: string[] = []

export const fixtureMisses = () => misses.slice()

const json = (body: unknown, paginated = false, status = 200) => {
  const headers = new Headers({ 'content-type': 'application/json' })
  if (paginated) {
    headers.set('X-Nuon-Page-Next', 'false')
    headers.set('X-Nuon-Page-Offset', '0')
    headers.set('X-Nuon-Page-Limit', '20')
  }
  return new Response(JSON.stringify(body), { status, headers })
}

class StoryEventSource extends EventTarget {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSED = 2
  static instances = new Set<StoryEventSource>()
  readyState = StoryEventSource.OPEN
  onopen: ((event: Event) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  url: string
  timers: ReturnType<typeof setTimeout>[] = []

  constructor(url: string) {
    super()
    this.url = url
    StoryEventSource.instances.add(this)
    queueMicrotask(() => {
      if (this.readyState === StoryEventSource.CLOSED) return
      this.onopen?.(new Event('open'))
      this.refresh()
    })
  }

  refresh() {
    this.timers.forEach(clearTimeout)
    const reply = current?.(new URL(this.url, window.location.origin), undefined)
    this.timers = (reply?.events ?? []).map(({ name, data, delay = 0 }) =>
      setTimeout(() => {
        if (this.readyState === StoryEventSource.CLOSED) return
        const event = new MessageEvent(name, {
          data: typeof data === 'string' ? data : JSON.stringify(data),
        })
        if (name === 'message') this.onmessage?.(event)
        this.dispatchEvent(event)
      }, delay)
    )
  }

  close() {
    this.timers.forEach(clearTimeout)
    this.readyState = StoryEventSource.CLOSED
    StoryEventSource.instances.delete(this)
  }
}

export const refreshInstallFixture = () => {
  StoryEventSource.instances.forEach((source) => source.refresh())
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
      return json({ api: { version: '0.0.0', git_ref: 'dev' }, ui: { version: '0.0.0' } })
    }
    if (url.pathname.startsWith('/v1/') && current) {
      const reply = current(url, init)
      if (reply?.pending) return new Promise(() => {})
      if (reply) return json(reply.body, reply.paginated, reply.status ?? 200)
      misses.push(`${init?.method ?? 'GET'} ${url.pathname}${url.search}`)
      return json({ error: 'fixture miss', description: url.pathname }, false, 404)
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
  StoryEventSource.instances.forEach((source) => source.close())
  if (restoreFetch) {
    window.fetch = restoreFetch
    restoreFetch = null
  }
  if (restoreEventSource) {
    window.EventSource = restoreEventSource
    restoreEventSource = null
  }
}
