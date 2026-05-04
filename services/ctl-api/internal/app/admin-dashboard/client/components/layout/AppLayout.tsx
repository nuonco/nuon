import { Outlet, Link, useLocation } from 'react-router'
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
      { path: '/runners', label: 'Runners' },
      { path: '/runners/all', label: 'All runners' },
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

export const AppLayout = () => {
  const location = useLocation()

  const isActive = (item: { path: string; exact?: boolean }) => {
    if (item.exact) return location.pathname === item.path
    return location.pathname === item.path || location.pathname.startsWith(item.path + '/')
  }

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-56 flex-shrink-0 border-r border-gray-200 bg-white">
        <div className="sticky top-0 flex h-screen flex-col overflow-y-auto">
          <div className="flex h-12 items-center px-4 border-b border-gray-200">
            <Link to="/" className="text-sm font-semibold tracking-tight text-primary-700">
              Nuon Admin
            </Link>
          </div>
          <nav className="flex-1 px-2 py-3 space-y-4">
            {navGroups.map((group, gi) => (
              <div key={gi}>
                {group.label && (
                  <p className="px-2 pb-1 text-[10px] font-semibold uppercase tracking-wider text-gray-400">
                    {group.label}
                  </p>
                )}
                <div className="space-y-0.5">
                  {group.items.map((item: { path: string; label: string; exact?: boolean; external?: boolean }) => {
                    const className = cn(
                      'block rounded-md px-2 py-1.5 text-[13px] font-medium transition-colors duration-100',
                      isActive(item)
                        ? 'bg-primary-50 text-primary-700'
                        : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900',
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
          <div className="mx-auto max-w-6xl">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
