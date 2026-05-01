import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getAccounts } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

const ACCOUNT_TYPE_OPTIONS = [
  { label: 'All', value: '' },
  { label: 'Nuon', value: 'nuon' },
  { label: 'User', value: 'user' },
]

export const AccountsList = () => {
  const [search, setSearch] = useState('')
  const [accountType, setAccountType] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['accounts', search, accountType, page],
    queryFn: () => getAccounts({ search, account_type: accountType || undefined, page }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load accounts'} />

  const { accounts = [], total_pages = 1 } = data || {}

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Accounts</h1>

      <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search accounts..." />
        </div>
        <div className="flex gap-2">
          {ACCOUNT_TYPE_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => { setAccountType(opt.value); setPage(1) }}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                accountType === opt.value
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Roles</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {accounts.map((account) => (
              <tr key={account.id} className="hover:bg-gray-50">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/accounts/${account.id}`} className="text-primary-600 hover:text-primary-800 font-medium">
                    {account.email}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(account.id)}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Badge>{account.account_type}</Badge>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{account.roles?.length ?? 0}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(account.created_at)}</td>
              </tr>
            ))}
            {accounts.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500">No accounts found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
