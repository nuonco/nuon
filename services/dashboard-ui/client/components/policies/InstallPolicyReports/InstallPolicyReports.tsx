import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PolicyReportPanel } from '@/components/policies/PolicyReportPanel'
import {
  reportSubject,
  type IPolicyReportRow,
} from '@/components/policies/PolicyReportsTable'
import type { TPolicyReport } from '@/types'

const OWNER_TYPE_LABELS: Record<
  string,
  { label: string; theme: 'info' | 'brand' | 'neutral' }
> = {
  install_deploys: { label: 'Deploy', theme: 'info' },
  install_sandbox_runs: { label: 'Sandbox', theme: 'brand' },
  component_builds: { label: 'Build', theme: 'neutral' },
}

const ViolationCounts = ({ report }: { report?: TPolicyReport }) => {
  const denyCount = report?.deny_count ?? 0
  const warnCount = report?.warn_count ?? 0

  if (!denyCount && !warnCount) {
    return (
      <Text variant="subtext" theme="neutral">
        None
      </Text>
    )
  }

  return (
    <span className="flex items-center gap-2">
      {denyCount ? (
        <Badge size="sm" theme="error">
          {denyCount} denied
        </Badge>
      ) : null}
      {warnCount ? (
        <Badge size="sm" theme="warn">
          {warnCount} warning{warnCount === 1 ? '' : 's'}
        </Badge>
      ) : null}
    </span>
  )
}

export interface IInstallPolicyReports {
  filterActions?: React.ReactNode
  filtered?: boolean
  loading?: boolean
  orgId: string
  policyNameMap: Map<string, string>
  rows: IPolicyReportRow[]
}

const loadingRows = (count: number): IPolicyReportRow[] =>
  Array.from({ length: count }, (_, index) => ({
    key: `loading-${index}`,
    report: {} as TPolicyReport,
    history: [],
  }))

export const InstallPolicyReports = ({
  filterActions,
  filtered = false,
  loading = false,
  orgId,
  policyNameMap,
  rows,
}: IInstallPolicyReports) => {
  const isLoading = loading && !rows.length
  const items = isLoading ? loadingRows(3) : rows

  return (
    <div className="flex flex-col gap-4 md:gap-6">
      {filterActions ? (
        <div className="flex flex-row flex-wrap items-center justify-end gap-4">
          {filterActions}
        </div>
      ) : null}

      {items.length ? (
        <div className="flex flex-col gap-4">
          {items.map(({ key, report, history }) => {
            const ownerType = report?.owner_type ?? ''
            const ownerMeta = OWNER_TYPE_LABELS[ownerType]

            return (
              <Card key={key} className="!p-4 !gap-4">
                <div className="flex items-start justify-between gap-3 flex-wrap">
                  <div className="flex flex-col gap-1.5 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap min-w-0">
                      <Icon
                        variant="ShieldCheckIcon"
                        size={16}
                        className="text-cool-grey-400 shrink-0"
                      />
                      <Text
                        variant="body"
                        weight="stronger"
                        role="heading"
                        level={3}
                        loading={isLoading}
                        loadingWidth={20}
                      >
                        {reportSubject(report)}
                      </Text>
                      {isLoading ? null : ownerMeta ? (
                        <Badge size="sm" theme={ownerMeta.theme}>
                          {ownerMeta.label}
                        </Badge>
                      ) : ownerType ? (
                        <Badge size="sm" theme="neutral">
                          {ownerType}
                        </Badge>
                      ) : null}
                    </div>
                    <ID loading={isLoading} loadingWidth={24}>
                      {report?.id ?? ''}
                    </ID>
                  </div>
                  {isLoading ? null : (
                    <PolicyReportPanel
                      report={report}
                      history={history}
                      orgId={orgId}
                      policyNameMap={policyNameMap}
                      panelKey={`policy-report-${report?.id}`}
                      triggerButton={{
                        variant: 'secondary',
                        size: 'sm',
                        children: 'View report',
                      }}
                    />
                  )}
                </div>

                <div className="flex flex-col gap-4 border-t pt-4">
                  <Text variant="subtext" weight="strong" theme="neutral">
                    Latest evaluation
                  </Text>
                  <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
                    <LabeledStatus
                      label="Status"
                      loading={isLoading}
                      statusProps={{ status: report?.status?.status }}
                      tooltipProps={{
                        tipContent: report?.status?.status_human_description,
                        position: 'bottom',
                      }}
                    />
                    <LabeledValue label="Evaluated" loading={isLoading}>
                      <Time
                        variant="subtext"
                        time={report?.evaluated_at ?? ''}
                        format="relative"
                      />
                    </LabeledValue>
                    <LabeledValue label="Violations" loading={isLoading}>
                      <ViolationCounts report={report} />
                    </LabeledValue>
                    <LabeledValue label="Evaluations" loading={isLoading}>
                      <Text variant="subtext" family="mono">
                        {history.length + 1}
                      </Text>
                    </LabeledValue>
                  </div>
                </div>
              </Card>
            )
          })}
        </div>
      ) : filtered ? (
        <EmptyState
          variant="policy"
          emptyTitle="No matching reports"
          emptyMessage="No reports match the current filters."
        />
      ) : (
        <EmptyState
          variant="policy"
          emptyTitle="No evaluations yet"
          emptyMessage="Evaluations appear here once a deploy or sandbox run triggers a policy check."
        />
      )}
    </div>
  )
}
