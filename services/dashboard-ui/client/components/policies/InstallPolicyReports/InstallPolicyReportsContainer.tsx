import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { PolicyReportsFilter } from '@/components/policies/PolicyReportsFilter'
import {
  PolicyReportsTable,
  groupPolicyReports,
} from '@/components/policies/PolicyReportsTable'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallPolicyReports, getAppPoliciesConfigs } from '@/lib'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'
import { InstallPolicyReports } from './InstallPolicyReports'

export const InstallPolicyReportsContainer = ({
  variant = 'table',
}: {
  variant?: 'table' | 'cards'
} = {}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const status = searchParams.get('status') as TPolicyReportStatus | null
  const ownerType = searchParams.get(
    'owner_type'
  ) as TPolicyReportOwnerType | null

  const { data: reportsResult, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'install-policy-reports',
      org?.id,
      install?.id,
      status,
      ownerType,
    ],
    queryFn: () =>
      getInstallPolicyReports({
        orgId: org.id,
        installId: install.id,
        status: status || undefined,
        ownerType: ownerType || undefined,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const { data: policiesConfigs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-policies-configs', org?.id, install?.app_id],
    queryFn: () =>
      getAppPoliciesConfigs({
        orgId: org.id,
        appId: install.app_id!,
      }),
    enabled: !!org?.id && !!install?.app_id,
  })

  const policyNameMap = useMemo(() => {
    const map = new Map<string, string>()
    if (!policiesConfigs) return map
    for (const config of policiesConfigs) {
      for (const policy of config.policies ?? []) {
        if (policy.id && policy.name) {
          map.set(policy.id, policy.name)
        }
      }
    }
    return map
  }, [policiesConfigs])

  const reports = reportsResult ?? []
  const rows = useMemo(() => groupPolicyReports(reports), [reports])

  if (variant === 'cards') {
    return (
      <InstallPolicyReports
        rows={rows}
        orgId={org?.id ?? ''}
        policyNameMap={policyNameMap}
        loading={isLoading}
        filtered={!!status || !!ownerType}
        filterActions={<PolicyReportsFilter />}
      />
    )
  }

  return (
    <PolicyReportsTable
      reports={reports}
      orgId={org?.id ?? ''}
      policyNameMap={policyNameMap}
      isLoading={isLoading}
      currentStatus={status || undefined}
      currentOwnerType={ownerType || undefined}
    />
  )
}
