import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getOrgDetail, getOrgStatus, getOrgInstalls, updateOrgTags, removeOrgTag, addSupportUsers, migrateOrgQueues } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const OrgDetail = () => {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [installsPage, setInstallsPage] = useState(1)
  const [newTag, setNewTag] = useState('')
  const [supportEmail, setSupportEmail] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['org', id],
    queryFn: () => getOrgDetail(id!),
    enabled: !!id,
  })

  const { data: statusData } = useQuery({
    queryKey: ['org-status', id],
    queryFn: () => getOrgStatus(id!),
    enabled: !!id,
    refetchInterval: 20000,
  })

  const { data: installsData } = useQuery({
    queryKey: ['org-installs', id, installsPage],
    queryFn: () => getOrgInstalls(id!, { page: installsPage }),
    enabled: !!id,
  })

  const tagsMutation = useMutation({
    mutationFn: (tags: string[]) => updateOrgTags(id!, tags),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const removeTagMutation = useMutation({
    mutationFn: (tag: string) => removeOrgTag(id!, tag),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const supportMutation = useMutation({
    mutationFn: (emails: string[]) => addSupportUsers(id!, emails),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org', id] })
      setSupportEmail('')
    },
  })

  const migrateMutation = useMutation({
    mutationFn: () => migrateOrgQueues(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load organization'} />
  if (!data) return null

  const { org, support_users = [] } = data
  const installs = installsData?.installs || []

  const handleAddTag = () => {
    if (!newTag.trim()) return
    const current = org.tags || []
    tagsMutation.mutate([...current, newTag.trim()])
    setNewTag('')
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">{org.name}</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{org.id}</p>
        <div className="mt-2 flex items-center gap-2">
          {statusData && (
            <Badge variant="status" status={statusData.status}>
              {statusData.status}
            </Badge>
          )}
          {statusData?.status_description && (
            <span className="text-xs text-gray-500">{statusData.status_description}</span>
          )}
        </div>
      </div>

      {/* Tags */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Tags</h2>
        <div className="mt-2 flex flex-wrap gap-2">
          {(org.tags || []).map((tag) => (
            <span key={tag} className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-700">
              {tag}
              <button
                onClick={() => removeTagMutation.mutate(tag)}
                className="ml-1 text-gray-400 hover:text-red-500"
                disabled={removeTagMutation.isPending}
              >
                x
              </button>
            </span>
          ))}
        </div>
        <div className="mt-2 flex gap-2">
          <input
            type="text"
            value={newTag}
            onChange={(e) => setNewTag(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddTag()}
            placeholder="Add tag..."
            className="block w-48 rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
          <button
            onClick={handleAddTag}
            disabled={tagsMutation.isPending}
            className="rounded-md bg-primary-600 px-3 py-1 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
          >
            Add
          </button>
        </div>
      </div>

      {/* Support Users */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Support Users</h2>
        {support_users.length > 0 ? (
          <ul className="mt-2 space-y-1">
            {support_users.map((email) => (
              <li key={email} className="text-sm text-gray-700">{email}</li>
            ))}
          </ul>
        ) : (
          <p className="mt-2 text-sm text-gray-500">No support users</p>
        )}
        <div className="mt-2 flex gap-2">
          <input
            type="email"
            value={supportEmail}
            onChange={(e) => setSupportEmail(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && supportEmail.trim() && supportMutation.mutate([supportEmail.trim()])}
            placeholder="Email address..."
            className="block w-64 rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
          <button
            onClick={() => supportEmail.trim() && supportMutation.mutate([supportEmail.trim()])}
            disabled={supportMutation.isPending}
            className="rounded-md bg-primary-600 px-3 py-1 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
          >
            Add
          </button>
        </div>
      </div>

      {/* Actions */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Actions</h2>
        <div className="mt-2">
          <button
            onClick={() => migrateMutation.mutate()}
            disabled={migrateMutation.isPending}
            className="rounded-md bg-yellow-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-yellow-700 disabled:opacity-50"
          >
            {migrateMutation.isPending ? 'Migrating...' : 'Migrate Queues'}
          </button>
          {migrateMutation.isSuccess && (
            <span className="ml-2 text-sm text-green-600">Done</span>
          )}
        </div>
      </div>

      {/* Installs */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Installs</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {installs.map((install: any) => (
                <tr key={install.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/installs/${install.id}`} className="text-primary-600 hover:text-primary-800 font-medium">
                      {install.name || truncateId(install.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(install.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={install.status}>{install.status}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(install.created_at)}</td>
                </tr>
              ))}
              {installs.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No installs found</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {installsData && (
          <Pagination page={installsPage} totalPages={installsData.total_pages} onPageChange={setInstallsPage} />
        )}
      </div>
    </div>
  )
}
