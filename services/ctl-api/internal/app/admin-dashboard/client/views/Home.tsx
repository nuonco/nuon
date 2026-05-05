import { Link } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { triggerPromotion, triggerSeed } from '@/lib/admin-api'

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

export const Home = () => {
  const promoteMutation = useMutation({
    mutationFn: triggerPromotion,
  })

  const seedMutation = useMutation({
    mutationFn: triggerSeed,
  })

  return (
    <div>
      <h1 className="page-heading">Admin dashboard</h1>
      <p className="page-subheading">Internal operations dashboard for the Nuon platform</p>

      <div className="mt-6 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-4 shadow-sm">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Global actions</h2>
        <div className="mt-3 flex items-center gap-3">
          <button
            onClick={() => {
              if (confirm('Are you sure you want to trigger a promotion? This will restart all org event loops and runners.')) {
                promoteMutation.mutate()
              }
            }}
            disabled={promoteMutation.isPending}
            className="rounded-md bg-primary-600 dark:bg-primary-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600 disabled:opacity-50"
          >
            {promoteMutation.isPending ? 'Promoting...' : 'Promote'}
          </button>
          <button
            onClick={() => {
              if (confirm('Are you sure you want to trigger a seed? This will terminate event loops and re-seed.')) {
                seedMutation.mutate()
              }
            }}
            disabled={seedMutation.isPending}
            className="rounded-md bg-yellow-600 dark:bg-yellow-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-yellow-700 dark:hover:bg-yellow-600 disabled:opacity-50"
          >
            {seedMutation.isPending ? 'Seeding...' : 'Seed'}
          </button>
          {promoteMutation.isSuccess && (
            <span className="text-sm text-green-600 dark:text-green-400">
              Promoted with tag: {promoteMutation.data.tag}
            </span>
          )}
          {promoteMutation.isError && (
            <span className="text-sm text-red-600 dark:text-red-400">Failed to promote</span>
          )}
          {seedMutation.isSuccess && (
            <span className="text-sm text-green-600 dark:text-green-400">Seed triggered</span>
          )}
          {seedMutation.isError && (
            <span className="text-sm text-red-600 dark:text-red-400">Failed to seed</span>
          )}
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {sections.map((s) => (
          <Link
            key={s.path}
            to={s.path}
            className="group rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-4 shadow-sm transition-all duration-150 hover:border-primary-200 dark:hover:border-primary-800 hover:shadow-md"
          >
            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 group-hover:text-primary-700 dark:group-hover:text-primary-300">{s.title}</h3>
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400 leading-relaxed">{s.description}</p>
          </Link>
        ))}
      </div>
    </div>
  )
}
