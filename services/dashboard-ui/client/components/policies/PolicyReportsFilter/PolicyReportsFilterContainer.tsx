import { useCallback } from 'react'
import { useLocation, useNavigate, useSearchParams } from 'react-router'
import { PolicyReportsFilter } from './PolicyReportsFilter'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'

export function PolicyReportsFilterContainer() {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const [searchParams] = useSearchParams()

  const currentStatus = (searchParams.get('status') as TPolicyReportStatus) || undefined
  const currentOwnerType = (searchParams.get('owner_type') as TPolicyReportOwnerType) || undefined

  const updateParam = useCallback(
    (key: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString())
      if (value) {
        params.set(key, value)
      } else {
        params.delete(key)
      }
      navigate(`${pathname}?${params.toString()}`, { replace: true })
    },
    [navigate, pathname, searchParams]
  )

  const handleStatusChange = useCallback(
    (value: string) => updateParam('status', value),
    [updateParam]
  )

  const handleTypeChange = useCallback(
    (value: string) => updateParam('owner_type', value),
    [updateParam]
  )

  const handleClearFilters = useCallback(() => {
    const params = new URLSearchParams(searchParams.toString())
    params.delete('status')
    params.delete('owner_type')
    navigate(`${pathname}?${params.toString()}`, { replace: true })
  }, [navigate, pathname, searchParams])

  return (
    <PolicyReportsFilter
      currentStatus={currentStatus}
      currentOwnerType={currentOwnerType}
      onStatusChange={handleStatusChange}
      onTypeChange={handleTypeChange}
      onClearFilters={handleClearFilters}
    />
  )
}
