import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { PolicyReportsFilter } from '@/components/policies/PolicyReportsFilter'
import { PolicyReportsTable } from '@/components/policies/PolicyReportsTable'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallPolicyReports, getAppPoliciesConfigs } from '@/lib'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'

export const Policies = () => {
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

  return (
    <PageSection>
      <PageTitle segments={['Policies', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/policies`,
            text: 'Policies',
          },
        ]}
      />
      <SectionHeader
        title="Policy reports"
        description="Latest policy evaluation for each component in this install."
        actions={<PolicyReportsFilter />}
      />

      <PolicyReportsTable
        reports={reports}
        orgId={org?.id ?? ''}
        policyNameMap={policyNameMap}
        isLoading={isLoading}
        currentStatus={status || undefined}
        currentOwnerType={ownerType || undefined}
      />
    </PageSection>
  )
}
