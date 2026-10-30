import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import {
  COLLECTION_VIEW_MODES,
  COLLECTION_VIEW_STORAGE_KEY,
  CollectionViewToggle,
  type TCollectionView,
} from '@/components/common/CollectionViewToggle'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { LINK_COLOR_CLASSES } from '@/components/common/Link'
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
import { useStoredViewMode } from '@/hooks/use-stored-view-mode'
import type { TPolicyReport } from '@/types'
import { cn } from '@/utils/classnames'

const TITLE_TRIGGER_CLASS = cn(
  '!p-0 !h-auto !border-none !rounded-none !bg-transparent hover:!bg-transparent active:!bg-transparent focus:!shadow-none !font-[inherit] !text-[length:inherit] !leading-[inherit]',
  LINK_COLOR_CLASSES
)

const FindingCounts = ({ report }: { report?: TPolicyReport }) => {
  const denyCount = report?.deny_count ?? 0
  const warnCount = report?.warn_count ?? 0
  const skipCount = Number(report?.status?.metadata?.skip_count ?? 0)

  if (!denyCount && !warnCount && !skipCount) {
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
      {skipCount ? (
        <Badge size="sm" theme="neutral">
          {skipCount} skip finding{skipCount === 1 ? '' : 's'}
        </Badge>
      ) : null}
    </span>
  )
}

export interface IInstallPolicyReports {
  filterActions?: React.ReactNode
  filtered?: boolean
  loading?: boolean
  onViewChange?: (view: TCollectionView) => void
  orgId: string
  policyNameMap: Map<string, string>
  rows: IPolicyReportRow[]
  view?: TCollectionView
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
  onViewChange,
  orgId,
  policyNameMap,
  rows,
  view: viewProp,
}: IInstallPolicyReports) => {
  const [storedView, setStoredView] = useStoredViewMode<TCollectionView>(
    COLLECTION_VIEW_STORAGE_KEY,
    COLLECTION_VIEW_MODES,
    'list'
  )
  const view = viewProp ?? storedView
  const setView = (next: TCollectionView) => {
    onViewChange?.(next)
    if (viewProp === undefined) setStoredView(next)
  }
  const isGrid = view === 'grid'
  const isLoading = loading && !rows.length
  const items = isLoading ? loadingRows(3) : rows

  return (
    <div className="flex flex-col gap-4 md:gap-6">
      <div className="flex flex-row flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap items-center gap-4 w-full md:w-fit">
          {filterActions}
        </div>
        <div className="flex items-center gap-3 ml-auto">
          <CollectionViewToggle value={view} onChange={setView} />
        </div>
      </div>

      {items.length ? (
        <div
          className={cn(
            'flex flex-col gap-4',
            isGrid && 'md:grid md:grid-cols-2'
          )}
        >
          {items.map(({ key, report, history }) => {
            return (
              <Card
                key={key}
                className={cn('!p-4 !gap-4', isGrid && 'md:h-full')}
              >
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
                        {isLoading ? (
                          reportSubject(report)
                        ) : (
                          <PolicyReportPanel
                            report={report}
                            history={history}
                            orgId={orgId}
                            policyNameMap={policyNameMap}
                            panelKey={`policy-report-${report?.id}`}
                            triggerButton={{
                              variant: 'ghost',
                              className: TITLE_TRIGGER_CLASS,
                              children: reportSubject(report),
                            }}
                          />
                        )}
                      </Text>
                    </div>
                    <ID loading={isLoading} loadingWidth={24}>
                      {report?.id ?? ''}
                    </ID>
                  </div>
                </div>

                <div className="flex flex-col gap-4 border-t pt-4">
                  <Text variant="subtext" weight="strong" theme="neutral">
                    Latest evaluation
                  </Text>
                  <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
                    {report?.status?.metadata?.resolved_action ===
                    'skip_apply' ? (
                      <LabeledValue label="Status" loading={isLoading}>
                        <Badge theme="neutral" size="sm">
                          Skip apply
                        </Badge>
                      </LabeledValue>
                    ) : (
                      <LabeledStatus
                        label="Status"
                        loading={isLoading}
                        statusProps={{ status: report?.status?.status }}
                        tooltipProps={{
                          tipContent: report?.status?.status_human_description,
                          position: 'bottom',
                        }}
                      />
                    )}
                    <LabeledValue label="Evaluated" loading={isLoading}>
                      <Time
                        variant="subtext"
                        time={report?.evaluated_at ?? ''}
                        format="relative"
                      />
                    </LabeledValue>
                    <LabeledValue label="Findings" loading={isLoading}>
                      <FindingCounts report={report} />
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
