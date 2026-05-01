import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getOrgs } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

export const OrgsList = () => {
  const [search, setSearch] = useState('')
  const [selectedTags, setSelectedTags] = useState<string[]>([])
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['orgs', search, selectedTags, page],
    queryFn: () => getOrgs({ search, tag: selectedTags, page }),
    refetchInterval: 20000,
  })

  const toggleTag = (tag: string) => {
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag],
    )
    setPage(1)
  }

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load organizations'} />

  const { orgs = [], all_tags = [], total_pages = 1 } = data || {}

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Organizations</h1>

      <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-start">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search orgs..." />
        </div>

        {all_tags.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {all_tags.map((tag) => (
              <label key={tag} className="flex items-center gap-1 text-xs text-gray-700 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedTags.includes(tag)}
                  onChange={() => toggleTag(tag)}
                  className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                />
                {tag}
              </label>
            ))}
          </div>
        )}
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Tags</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Apps</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Installs</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {orgs.map((org) => (
              <tr key={org.id} className="hover:bg-gray-50">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/orgs/${org.id}`} className="text-primary-600 hover:text-primary-800 font-medium">
                    {org.name}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(org.id)}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  {org.status && (
                    <Badge variant="status" status={getStatus(org.status)}>{getStatus(org.status)}</Badge>
                  )}
                </td>
                <td className="px-4 py-3 text-sm">
                  <div className="flex flex-wrap gap-1">
                    {(org.tags || []).map((tag) => (
                      <Badge key={tag}>{tag}</Badge>
                    ))}
                  </div>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{org.app_count ?? '-'}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{org.install_count ?? '-'}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(org.created_at)}</td>
              </tr>
            ))}
            {orgs.length === 0 && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-sm text-gray-500">No organizations found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
