'use client'

import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { ChangeEvent, useCallback } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'

const STATUS_OPTIONS: {
  value: TPolicyReportStatus
  label: string
  theme: 'error' | 'warn' | 'success'
}[] = [
  { value: 'error', label: 'Denied', theme: 'error' },
  { value: 'warning', label: 'Warning', theme: 'warn' },
  { value: 'success', label: 'Passed', theme: 'success' },
]

const TYPE_OPTIONS: {
  value: TPolicyReportOwnerType
  label: string
  theme: 'info' | 'brand' | 'neutral'
}[] = [
  { value: 'install_deploys', label: 'Deploy', theme: 'info' },
  { value: 'install_sandbox_runs', label: 'Sandbox', theme: 'brand' },
  // Installs don't have reports where owner_type is 'component_builds', so this filter option is currently hidden.
  // { value: 'component_builds', label: 'Build', theme: 'neutral' },
]

interface IPolicyReportsFilter {
  currentStatus?: TPolicyReportStatus
  currentOwnerType?: TPolicyReportOwnerType
}

export function PolicyReportsFilter({
  currentStatus,
  currentOwnerType,
}: IPolicyReportsFilter) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const updateParam = useCallback(
    (key: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString())

      if (value) {
        params.set(key, value)
      } else {
        params.delete(key)
      }

      router.replace(`${pathname}?${params.toString()}`, { scroll: false })
    },
    [router, pathname, searchParams]
  )

  const handleStatusChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      updateParam('status', e.target.value)
    },
    [updateParam]
  )

  const handleTypeChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      updateParam('owner_type', e.target.value)
    },
    [updateParam]
  )

  const handleClearFilters = useCallback(() => {
    const params = new URLSearchParams(searchParams.toString())
    params.delete('status')
    params.delete('owner_type')
    router.replace(`${pathname}?${params.toString()}`, { scroll: false })
  }, [router, pathname, searchParams])

  const activeFilterCount =
    (currentStatus ? 1 : 0) + (currentOwnerType ? 1 : 0)

  return (
    <Dropdown
      alignment="right"
      buttonClassName="!p-2"
      buttonText={
        <>
          <Icon variant="FunnelSimple" />
          Filter{activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
        </>
      }
      id="policy-reports-filter"
      variant="ghost"
    >
      <Menu className="!p-0 !w-56">
        <form onReset={handleClearFilters}>
          <div className="p-2">
            <Text variant="label" theme="neutral" className="px-1 mb-1">
              Status
            </Text>
            <div className="flex flex-col gap-0.5">
              {STATUS_OPTIONS.map((option) => (
                <RadioInput
                  key={option.value}
                  checked={currentStatus === option.value}
                  labelProps={{
                    labelText: (
                      <Badge theme={option.theme} size="sm">
                        {option.label}
                      </Badge>
                    ),
                  }}
                  name="status"
                  onChange={handleStatusChange}
                  value={option.value}
                />
              ))}
            </div>
          </div>

          <hr />

          <div className="p-2">
            <Text variant="label" theme="neutral" className="px-1 mb-1">
              Type
            </Text>
            <div className="flex flex-col gap-0.5">
              {TYPE_OPTIONS.map((option) => (
                <RadioInput
                  key={option.value}
                  checked={currentOwnerType === option.value}
                  labelProps={{
                    labelText: (
                      <Badge theme={option.theme} size="sm">
                        {option.label}
                      </Badge>
                    ),
                  }}
                  name="owner_type"
                  onChange={handleTypeChange}
                  value={option.value}
                />
              ))}
            </div>
          </div>

          <hr />

          <div className="p-2">
            <Button
              className="w-full"
              isMenuButton
              type="reset"
              variant="ghost"
            >
              Clear filters
              <Icon variant="X" />
            </Button>
          </div>
        </form>
      </Menu>
    </Dropdown>
  )
}
