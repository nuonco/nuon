import { Outlet, Link, useLocation } from 'react-router'
import { useEffect, useState } from 'react'
import cn from 'classnames'

const navGroups = [
  {
    items: [
      { path: '/', label: 'Home', exact: true },
    ],
  },
  {
    label: 'Resources',
    items: [
      { path: '/orgs', label: 'Orgs' },
      { path: '/accounts', label: 'Accounts' },
      { path: '/installs', label: 'Installs' },
      { path: '/runners/all', label: 'Runners' },
      { path: '/runner-uptime', label: 'Runner uptime' },
    ],
  },
  {
    label: 'Queues & signals',
    items: [
      { path: '/queues', label: 'Queues' },
      { path: '/workflows', label: 'Workflows' },
      { path: '/queue-signals', label: 'Queue signals' },
      { path: '/in-flight-signals', label: 'In-flight' },
      { path: '/signal-catalog', label: 'Signal catalog' },
    ],
  },
  {
    label: 'System',
    items: [
      { path: '/log-streams', label: 'Log streams' },
      { path: '/labels', label: 'Labels' },
      { path: '/sandbox-mode', label: 'Sandbox mode' },
      { path: '/temporal-workers', label: 'Temporal workers' },
      { path: '/temporal-workflows', label: 'Temporal workflows' },
      { path: '/queries', label: 'Queries' },
      { path: '/query-catalog', label: 'Query catalog' },
      { path: '/api/livez', label: 'Health', external: true },
    ],
  },
]

function useDarkMode() {
  const [dark, setDark] = useState(() => {
    if (typeof window === 'undefined') return false
    const stored = localStorage.getItem('nuon-admin-dark')
    if (stored !== null) return stored === 'true'
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })

  useEffect(() => {
    const root = document.documentElement
    if (dark) {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
    localStorage.setItem('nuon-admin-dark', String(dark))
  }, [dark])

  return [dark, setDark] as const
}

export const AppLayout = () => {
  const location = useLocation()
  const [dark, setDark] = useDarkMode()

  const isActive = (item: { path: string; exact?: boolean }) => {
    if (item.exact) return location.pathname === item.path
    return location.pathname === item.path || location.pathname.startsWith(item.path + '/')
  }

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-56 flex-shrink-0 border-r border-gray-200 dark:border-gray-800">
        <div className="sticky top-0 flex h-screen flex-col overflow-y-auto">
          <div className="flex h-12 items-center justify-between px-4 border-b border-gray-200 dark:border-gray-800">
            <Link to="/" className="text-sm font-semibold tracking-tight text-primary-700 dark:text-primary-400">
              Nuon Admin
            </Link>
            <button
              onClick={() => setDark(!dark)}
              className="rounded p-1 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
              title={dark ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {dark ? (
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" />
                </svg>
              ) : (
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z" />
                </svg>
              )}
            </button>
          </div>
          <nav className="flex-1 px-2 py-3 space-y-4">
            {navGroups.map((group, gi) => (
              <div key={gi}>
                {group.label && (
                  <p className="px-2 pb-1 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">
                    {group.label}
                  </p>
                )}
                <div className="space-y-0.5">
                  {group.items.map((item: { path: string; label: string; exact?: boolean; external?: boolean }) => {
                    const className = cn(
                      'block rounded-md px-2 py-1.5 text-[13px] font-medium transition-colors duration-100',
                      isActive(item)
                        ? 'bg-primary-50 text-primary-700 dark:bg-primary-950 dark:text-primary-300'
                        : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-900 dark:hover:text-gray-100',
                    )
                    if (item.external) {
                      return (
                        <a key={item.path} href={item.path} target="_blank" rel="noopener noreferrer" className={className}>
                          {item.label}
                        </a>
                      )
                    }
                    return (
                      <Link key={item.path} to={item.path} className={className}>
                        {item.label}
                      </Link>
                    )
                  })}
                </div>
              </div>
            ))}
          </nav>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 min-w-0">
        <main className="p-6 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
