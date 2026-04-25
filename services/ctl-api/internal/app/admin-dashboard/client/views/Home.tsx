import { Link } from 'react-router'

const sections = [
  { path: '/orgs', title: 'Organizations', description: 'Manage organizations, tags, and support users' },
  { path: '/accounts', title: 'Accounts', description: 'View accounts, roles, and audit logs' },
  { path: '/installs', title: 'Installs', description: 'Monitor installs, deployments, and activity' },
  { path: '/runners', title: 'Runners', description: 'View runners and sandbox configurations' },
  { path: '/queues', title: 'Queues', description: 'Manage queues, emitters, and signals' },
  { path: '/workflows', title: 'Workflows', description: 'Inspect workflow executions and steps' },
  { path: '/log-streams', title: 'Log streams', description: 'View log stream data from ClickHouse' },
  { path: '/sandbox-mode', title: 'Sandbox mode', description: 'Configure sandbox runner jobs and signals' },
  { path: '/temporal-workers', title: 'Temporal workers', description: 'Monitor Temporal worker health' },
]

export const Home = () => (
  <div>
    <h1 className="text-2xl font-bold text-gray-900">Admin dashboard</h1>
    <p className="mt-1 text-sm text-gray-500">Internal operations dashboard for Nuon platform</p>
    <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {sections.map((s) => (
        <Link
          key={s.path}
          to={s.path}
          className="rounded-lg border border-gray-200 bg-white p-5 shadow-sm transition-shadow hover:shadow-md"
        >
          <h3 className="text-sm font-semibold text-gray-900">{s.title}</h3>
          <p className="mt-1 text-xs text-gray-500">{s.description}</p>
        </Link>
      ))}
    </div>
  </div>
)
