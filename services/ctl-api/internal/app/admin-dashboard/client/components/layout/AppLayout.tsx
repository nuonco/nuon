import { Outlet, Link, useLocation } from 'react-router'
import cn from 'classnames'

const navItems = [
  { path: '/', label: 'Home', exact: true },
  { path: '/orgs', label: 'Orgs' },
  { path: '/accounts', label: 'Accounts' },
  { path: '/installs', label: 'Installs' },
  { path: '/runners', label: 'Runners' },
  { path: '/queues', label: 'Queues' },
  { path: '/workflows', label: 'Workflows' },
  { path: '/queue-signals', label: 'Signals' },
  { path: '/in-flight-signals', label: 'In-flight' },
  { path: '/signal-catalog', label: 'Catalog' },
  { path: '/log-streams', label: 'Logs' },
  { path: '/labels', label: 'Labels' },
  { path: '/sandbox-mode', label: 'Sandbox' },
  { path: '/temporal-workers', label: 'Workers' },
  { path: '/temporal-workflows', label: 'Temporal' },
]

export const AppLayout = () => {
  const location = useLocation()

  const isActive = (item: typeof navItems[number]) => {
    if (item.exact) return location.pathname === item.path
    return location.pathname.startsWith(item.path)
  }

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-gray-900 text-white">
        <div className="mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex h-12 items-center justify-between">
            <Link to="/" className="text-lg font-bold tracking-tight">
              Nuon Admin
            </Link>
          </div>
        </div>
        <nav className="border-t border-gray-700">
          <div className="mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex space-x-1 overflow-x-auto py-1">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={cn(
                    'whitespace-nowrap rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                    isActive(item)
                      ? 'bg-gray-700 text-white'
                      : 'text-gray-300 hover:bg-gray-800 hover:text-white',
                  )}
                >
                  {item.label}
                </Link>
              ))}
            </div>
          </div>
        </nav>
      </header>
      <main className="flex-1 bg-gray-50">
        <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
