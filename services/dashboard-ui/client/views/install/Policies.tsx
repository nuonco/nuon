import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PolicyReportsTable } from '@/components/policies/PolicyReportsTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallPolicyReports, getAppPoliciesConfigs } from '@/lib'
import type { TPolicyReportOwnerType, TPolicyReportStatus } from '@/lib/ctl-api/installs/get-install-policy-reports'

export const Policies = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const status = searchParams.get('status') as TPolicyReportStatus | null
  const ownerType = searchParams.get('owner_type') as TPolicyReportOwnerType | null

  const { data: reportsResult, isLoading } = useQuery({
    queryKey: ['install-policy-reports', org?.id, install?.id, status, ownerType],
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
      <PageTitle title={`Policies | ${install?.name}`} />
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
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Policy reports
        </Text>
        <Text theme="neutral">
          View policy compliance reports for this install.
        </Text>
      </HeadingGroup>

      <div className="flex flex-auto">
        {isLoading ? (
          <div className="flex flex-col gap-3 w-full">
            {Array.from({ length: 3 }).map((_, i) => (
              <div
                key={i}
                className="h-32 rounded-md border bg-cool-grey-50 dark:bg-dark-grey-800 animate-pulse"
              />
            ))}
          </div>
        ) : (
          <PolicyReportsTable
            reports={reports}
            orgId={org?.id ?? ''}
            installId={install?.id ?? ''}
            policyNameMap={policyNameMap}
            currentStatus={status || undefined}
            currentOwnerType={ownerType || undefined}
          />
        )}
      </div>
    </PageSection>
  )
}
