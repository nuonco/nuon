import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getOrgDetail, addOrgLabels, removeOrgLabel, addSupportUsers, migrateOrgQueues, updateOrgTags } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { DotGraph } from '@/components/common/DotGraph'
import { formatDate, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

const ORG_TAG_OPTIONS = ['Trial', 'Customer', 'Inactive', 'Priority', 'Demo', 'Employee']

export const OrgDetail = () => {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [installsPage, setInstallsPage] = useState(1)
  const [labelKey, setLabelKey] = useState('')
  const [labelValue, setLabelValue] = useState('')
  const [editingTags, setEditingTags] = useState(false)
  const [tagDraft, setTagDraft] = useState<string[]>([])

  const { data, isLoading, error } = useQuery({
    queryKey: ['org', id, installsPage],
    queryFn: () => getOrgDetail(id!, { page: installsPage }),
    enabled: !!id,
    refetchInterval: 20000,
  })

  const addLabelMutation = useMutation({
    mutationFn: (labels: Record<string, string>) => addOrgLabels(id!, labels),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org', id] })
      setLabelKey('')
      setLabelValue('')
    },
  })

  const removeLabelMutation = useMutation({
    mutationFn: (key: string) => removeOrgLabel(id!, key),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const updateTagsMutation = useMutation({
    mutationFn: (tags: string[]) => updateOrgTags(id!, tags),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org', id] })
      setEditingTags(false)
    },
  })

  const supportMutation = useMutation({
    mutationFn: () => addSupportUsers(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const migrateMutation = useMutation({
    mutationFn: () => migrateOrgQueues(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load organization'} />
  if (!data) return null

  const { org, installs = [], recent_app, graph_dot, app_url, installs_total_pages = 1 } = data
  const orgLabels = org.labels || {}

  const handleAddLabel = () => {
    if (!labelKey.trim()) return
    addLabelMutation.mutate({ [labelKey.trim()]: labelValue.trim() })
  }

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-1 text-sm text-gray-500">
        <Link to="/orgs" className="text-primary-600 hover:text-primary-700">Orgs</Link>
        <span>/</span>
        <span className="text-gray-900 font-medium">{org.name}</span>
      </nav>

      {/* Header */}
      <div>
        <div className="flex items-center gap-3">
          <h1 className="page-heading">{org.name}</h1>
          {app_url && (
            <a
              href={`${app_url}/${org.id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700"
            >
              View Dashboard
            </a>
          )}
        </div>
        <p className="mt-1 text-sm text-gray-500 font-mono">{org.id}</p>
        <div className="mt-2 flex items-center gap-2">
          {org.status && (
            <Badge variant="status" status={getStatus(org.status)}>{getStatus(org.status)}</Badge>
          )}
          {org.status_description && (
            <span className="text-xs text-gray-500">{org.status_description}</span>
          )}
        </div>
        <div className="mt-2 text-xs text-gray-500 flex gap-4">
          <span>Created: {formatDate(org.created_at)}</span>
          <span>Updated: {formatDate(org.updated_at)}</span>
        </div>
      </div>

      {/* Tags */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-900">Tags</h2>
          {!editingTags ? (
            <button
              onClick={() => {
                setTagDraft(org.tags || [])
                setEditingTags(true)
              }}
              className="text-xs text-primary-600 hover:text-primary-700 font-medium"
            >
              Edit
            </button>
          ) : (
            <div className="flex items-center gap-2">
              <button
                onClick={() => setEditingTags(false)}
                className="text-xs text-gray-500 hover:text-gray-700"
                disabled={updateTagsMutation.isPending}
              >
                Cancel
              </button>
              <button
                onClick={() => updateTagsMutation.mutate(tagDraft)}
                disabled={updateTagsMutation.isPending}
                className="rounded-md bg-primary-600 px-2 py-1 text-xs font-medium text-white hover:bg-primary-700 disabled:opacity-50"
              >
                {updateTagsMutation.isPending ? 'Saving...' : 'Save'}
              </button>
            </div>
          )}
        </div>
        {!editingTags ? (
          <div className="mt-2 flex flex-wrap gap-2">
            {(org.tags || []).map((tag: string) => (
              <Badge key={tag}>{tag}</Badge>
            ))}
            {(org.tags || []).length === 0 && <span className="text-sm text-gray-500">No tags</span>}
          </div>
        ) : (
          <div className="mt-2 grid grid-cols-2 gap-1 sm:grid-cols-3">
            {ORG_TAG_OPTIONS.map((tag) => {
              const checked = tagDraft.includes(tag)
              return (
                <label
                  key={tag}
                  className="flex cursor-pointer items-center gap-2 rounded border border-transparent px-2 py-1.5 text-sm hover:border-primary-200 hover:bg-primary-50"
                >
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={(e) => {
                      setTagDraft((curr) =>
                        e.target.checked ? [...curr, tag] : curr.filter((t) => t !== tag),
                      )
                    }}
                    className="rounded border-gray-300"
                  />
                  <span>{tag}</span>
                </label>
              )
            })}
          </div>
        )}
      </div>

      {/* Labels (editable) */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Labels</h2>
        {Object.keys(orgLabels).length > 0 ? (
          <div className="mt-2 flex flex-wrap gap-2">
            {Object.entries(orgLabels).map(([key, value]) => (
              <span key={key} className="inline-flex items-center gap-1 rounded-md bg-blue-50 border border-blue-200 px-2 py-0.5 text-xs font-mono">
                <span className="text-blue-700">{key}</span>
                <span className="text-blue-400">=</span>
                <span className="text-blue-600">{String(value)}</span>
                <button
                  onClick={() => removeLabelMutation.mutate(key)}
                  disabled={removeLabelMutation.isPending}
                  className="ml-1 text-blue-400 hover:text-red-500"
                >
                  &times;
                </button>
              </span>
            ))}
          </div>
        ) : (
          <p className="mt-2 text-sm text-gray-500">No labels</p>
        )}
        <div className="mt-2 flex gap-2">
          <input
            type="text"
            value={labelKey}
            onChange={(e) => setLabelKey(e.target.value)}
            placeholder="Key"
            className="block w-32 rounded-md border-0 py-1.5 px-2.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400"
          />
          <input
            type="text"
            value={labelValue}
            onChange={(e) => setLabelValue(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddLabel()}
            placeholder="Value"
            className="block w-40 rounded-md border-0 py-1.5 px-2.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400"
          />
          <button
            onClick={handleAddLabel}
            disabled={addLabelMutation.isPending}
            className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
          >
            Add
          </button>
        </div>
      </div>

      {/* Most Recent App */}
      {recent_app && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Most Recent App</h2>
          <div className="mt-2 space-y-1">
            <p className="text-sm text-gray-900 font-medium">{recent_app.name}</p>
            <p className="text-xs text-gray-500 font-mono">{recent_app.id}</p>
            {recent_app.status && (
              <Badge variant="status" status={getStatus(recent_app.status)}>{getStatus(recent_app.status)}</Badge>
            )}
          </div>
        </div>
      )}

      {/* Component Graph */}
      {graph_dot && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Component dependency graph</h2>
          <div className="mt-2">
            <DotGraph dot={graph_dot} height="28rem" />
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Actions</h2>
        <div className="mt-2 flex gap-3">
          <div>
            <button
              onClick={() => supportMutation.mutate()}
              disabled={supportMutation.isPending}
              className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
            >
              {supportMutation.isPending ? 'Adding...' : 'Add support users'}
            </button>
            {supportMutation.isSuccess && <span className="ml-2 text-sm text-green-600">Done</span>}
            {supportMutation.isError && <span className="ml-2 text-sm text-red-600">Failed</span>}
          </div>
          <div>
            <button
              onClick={() => migrateMutation.mutate()}
              disabled={migrateMutation.isPending}
              className="rounded-md bg-orange-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50"
            >
              {migrateMutation.isPending ? 'Migrating...' : 'Migrate queues'}
            </button>
            {migrateMutation.isSuccess && <span className="ml-2 text-sm text-green-600">Migration started</span>}
            {migrateMutation.isError && <span className="ml-2 text-sm text-red-600">Failed</span>}
          </div>
        </div>
      </div>

      {/* Installs */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Installs</h2>
        <div className="mt-2 table-card">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>ID</th>
                <th>App</th>
                <th>Runner</th>
                <th>Sandbox</th>
                <th>Components</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {installs.map((install: any) => {
                const isDeleted = !!install.deleted_at
                const runnerStatus = getStatus(install.runner_status)
                const sandboxStatus = getStatus(install.sandbox_status)
                const componentStatus = getStatus(install.composite_component_status)
                return (
                  <tr key={install.id} className={isDeleted ? 'opacity-50' : ''}>
                    <td>
                      <div className="flex items-center gap-2">
                        <Link to={`/installs/${install.id}`} className="text-primary-600 hover:text-primary-700 font-medium">
                          {install.name || truncateId(install.id)}
                        </Link>
                        {isDeleted && <Badge variant="status" status="error">Deleted</Badge>}
                      </div>
                    </td>
                    <td className="text-gray-500 font-mono text-xs">{truncateId(install.id)}</td>
                    <td className="text-sm text-gray-700">
                      {install.app?.name || truncateId(install.app_id || '')}
                    </td>
                    <td>
                      {runnerStatus && (
                        <Badge variant="status" status={runnerStatus}>{runnerStatus}</Badge>
                      )}
                    </td>
                    <td>
                      {sandboxStatus && (
                        <Badge variant="status" status={sandboxStatus}>{sandboxStatus}</Badge>
                      )}
                    </td>
                    <td>
                      {componentStatus && (
                        <Badge variant="status" status={componentStatus}>{componentStatus}</Badge>
                      )}
                    </td>
                    <td className="text-gray-500">{formatDate(install.created_at)}</td>
                  </tr>
                )
              })}
              {installs.length === 0 && (
                <tr><td colSpan={7} className="text-center text-gray-500 py-6">No installs found</td></tr>
              )}
            </tbody>
          </table>
        </div>
        <Pagination page={installsPage} totalPages={installs_total_pages} onPageChange={setInstallsPage} />
      </div>
    </div>
  )
}
