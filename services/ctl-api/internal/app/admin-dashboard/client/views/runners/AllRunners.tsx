import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Link } from 'react-router'
import { getAllRunners } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { truncateId } from '@/utils/format'
import type { TAllRunnerView } from '@/types/admin.types'

const COLORS = [
  '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
  '#ec4899', '#06b6d4', '#84cc16', '#f97316', '#6366f1',
]

const PieChart = ({ data, title }: { data: { label: string; value: number }[]; title: string }) => {
  const total = data.reduce((sum, d) => sum + d.value, 0)
  if (total === 0) return null

  let cumulative = 0
  const slices = data.map((d, i) => {
    const start = cumulative
    cumulative += d.value
    const startAngle = (start / total) * 360
    const endAngle = (cumulative / total) * 360
    return { ...d, startAngle, endAngle, color: COLORS[i % COLORS.length] }
  })

  const toRad = (deg: number) => (deg - 90) * (Math.PI / 180)
  const r = 60
  const cx = 70
  const cy = 70

  return (
    <div className="flex flex-col items-center gap-2">
      <h3 className="text-sm font-medium text-gray-700">{title}</h3>
      <div className="flex items-center gap-4">
        <svg width={140} height={140} viewBox="0 0 140 140">
          {slices.map((s, i) => {
            const largeArc = s.endAngle - s.startAngle > 180 ? 1 : 0
            const x1 = cx + r * Math.cos(toRad(s.startAngle))
            const y1 = cy + r * Math.sin(toRad(s.startAngle))
            const x2 = cx + r * Math.cos(toRad(s.endAngle))
            const y2 = cy + r * Math.sin(toRad(s.endAngle))

            if (data.length === 1) {
              return <circle key={i} cx={cx} cy={cy} r={r} fill={s.color} />
            }

            return (
              <path
                key={i}
                d={`M${cx},${cy} L${x1},${y1} A${r},${r} 0 ${largeArc},1 ${x2},${y2} Z`}
                fill={s.color}
              />
            )
          })}
        </svg>
        <div className="flex flex-col gap-1">
          {slices.map((s, i) => (
            <div key={i} className="flex items-center gap-2 text-xs">
              <div className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: s.color }} />
              <span className="text-gray-600">{s.label}: {s.value}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export const AllRunners = () => {
  const [orgId, setOrgId] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['all-runners', orgId],
    queryFn: () => getAllRunners({ org_id: orgId || undefined }),
    refetchInterval: 30000,
  })

  const runners = data?.runners || []
  const orgs = data?.orgs || []

  const stats = useMemo(() => {
    const groupTypeCounts: Record<string, number> = {}
    const versionCounts: Record<string, number> = {}
    const processTypeCounts: Record<string, number> = {}

    for (const rv of runners) {
      const gt = rv.group_type || 'unknown'
      groupTypeCounts[gt] = (groupTypeCounts[gt] || 0) + 1

      const v = rv.version || 'unknown'
      versionCounts[v] = (versionCounts[v] || 0) + 1

      const pt = rv.process_type || 'none'
      processTypeCounts[pt] = (processTypeCounts[pt] || 0) + 1
    }

    return {
      groupType: Object.entries(groupTypeCounts).map(([label, value]) => ({ label, value })),
      version: Object.entries(versionCounts).map(([label, value]) => ({ label, value })),
      processType: Object.entries(processTypeCounts).map(([label, value]) => ({ label, value })),
    }
  }, [runners])

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load runners'} />

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-gray-900">All runners</h1>
        <span className="text-sm text-gray-500">{runners.length} runners</span>
      </div>

      <div className="mt-4">
        <select
          value={orgId}
          onChange={(e) => setOrgId(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
        >
          <option value="">All organizations</option>
          {orgs.map((org) => (
            <option key={org.id} value={org.id}>{org.name}</option>
          ))}
        </select>
      </div>

      {runners.length > 0 && (
        <div className="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-3">
          <PieChart data={stats.groupType} title="Install vs org runners" />
          <PieChart data={stats.version} title="Versions" />
          <PieChart data={stats.processType} title="Process type (mng vs non-mng)" />
        </div>
      )}

      <div className="mt-6 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Org</th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Type</th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Version</th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Process</th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Owner</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {runners.map((rv) => (
              <RunnerRow key={rv.runner.id} rv={rv} />
            ))}
            {runners.length === 0 && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-sm text-gray-500">No runners found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

const RunnerRow = ({ rv }: { rv: TAllRunnerView }) => (
  <tr className="hover:bg-gray-50">
    <td className="whitespace-nowrap px-4 py-3 text-sm">
      <Link to={`/runners/${rv.runner.id}`} className="font-medium text-primary-600 hover:text-primary-800">
        {rv.runner.name || truncateId(rv.runner.id)}
      </Link>
    </td>
    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{rv.org_name || '-'}</td>
    <td className="whitespace-nowrap px-4 py-3 text-sm">
      <Badge variant="status" status={rv.group_type === 'install' ? 'info' : 'neutral'}>
        {rv.group_type || '-'}
      </Badge>
    </td>
    <td className="whitespace-nowrap px-4 py-3 text-sm">
      <Badge variant="status" status={rv.process_online ? 'online' : 'offline'}>
        {rv.process_online ? 'Online' : 'Offline'}
      </Badge>
    </td>
    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{rv.version || '-'}</td>
    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{rv.process_type || '-'}</td>
    <td className="whitespace-nowrap px-4 py-3 text-sm">
      {rv.install_id ? (
        <Link to={`/installs/${rv.install_id}`} className="text-primary-600 hover:text-primary-800">
          {rv.install_name || truncateId(rv.install_id)}
        </Link>
      ) : (
        <span className="text-gray-500">org</span>
      )}
    </td>
  </tr>
)
