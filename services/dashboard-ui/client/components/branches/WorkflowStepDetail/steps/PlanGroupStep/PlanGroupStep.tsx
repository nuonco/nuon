import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { Expand } from '@/components/common/Expand'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import {
  AppConfigDiff,
  type DiffSectionData,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'

export interface PlanInstallDiff {
  installId: string
  installName: string
  installLabels?: Record<string, string>
  sections: DiffSectionData[]
  summary: { added: number; removed: number; changed: number } | null
  isLoading?: boolean
}

interface IPlanGroupStep {
  installs: PlanInstallDiff[]
  groupName?: string
  labelColors?: Record<string, string>
  hasResponse: boolean
  responseType?: string
  showApproveBar: boolean
  isInProgress: boolean
  actions?: ReactNode
}

export const PlanGroupStep = ({
  installs,
  groupName,
  labelColors,
  hasResponse,
  responseType,
  showApproveBar,
  isInProgress: _isInProgress,
  actions,
}: IPlanGroupStep) => {
  return (
    <div className="flex flex-col gap-3">
      {hasResponse && (
        <Banner theme="success">
          <Text weight="strong">
            Plan {responseType === 'approve' ? 'approved' : responseType || 'responded'}
          </Text>
        </Banner>
      )}

      {showApproveBar && (
        <Banner className="@container" theme="warn">
          <div className="flex flex-col gap-2">
            <div className="flex flex-col">
              <Text weight="strong">Install group plan requires review</Text>
              <Text variant="subtext" theme="neutral">
                Review the changes below, then approve to deploy or skip this install group.
              </Text>
            </div>
            {actions && <div className="flex self-end gap-2">{actions}</div>}
          </div>
        </Banner>
      )}

      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 sm:px-6 py-4 border-b">
          <div className="flex items-center gap-3">
            <Icon variant="ListChecksIcon" size="16" />
            <Text variant="base" weight="strong">
              {groupName || 'Install group'}
            </Text>
            <Text variant="subtext" theme="neutral">
              {installs.length} {installs.length === 1 ? 'install' : 'installs'}
            </Text>
          </div>
        </div>

        <div className="flex flex-col divide-y">
          {installs.map((inst) => {
            const total =
              (inst.summary?.added ?? 0) +
              (inst.summary?.removed ?? 0) +
              (inst.summary?.changed ?? 0)
            const hasChanges = total > 0 && inst.sections.length > 0
            const labelEntries = inst.installLabels ? Object.entries(inst.installLabels) : []

            const heading = (
              <div className="flex items-center gap-3 w-full">
                <Text weight="strong">{inst.installName || inst.installId}</Text>
                {labelEntries.map(([k, v]) => (
                  <LabelBadge
                    key={k}
                    labelKey={k}
                    labelValue={v}
                    size="sm"
                    className="shrink-0"
                    customColor={labelColors?.[k]}
                  />
                ))}
                {inst.isLoading ? (
                  <Text variant="subtext" theme="neutral" className="ml-auto shrink-0">
                    Loading…
                  </Text>
                ) : (
                  <ChangeCountSummary
                    added={inst.summary?.added ?? 0}
                    updated={inst.summary?.changed ?? 0}
                    removed={inst.summary?.removed ?? 0}
                    emptyText="No changes"
                    className="ml-auto shrink-0"
                  />
                )}
              </div>
            )

            if (!hasChanges) {
              return (
                <div key={inst.installId} className="flex items-center gap-2 px-4 py-3">
                  {heading}
                  <Icon variant="CaretDownIcon" className="invisible shrink-0" aria-hidden />
                </div>
              )
            }

            return (
              <Expand
                key={inst.installId}
                id={`plan-install-${inst.installId}`}
                heading={heading}
                headerClassName="px-4 py-3"
              >
                <div className="px-4 py-4 border-t border-cool-grey-100 dark:border-dark-grey-800 bg-black/[0.015] dark:bg-white/[0.0075]">
                  <AppConfigDiff sections={inst.sections} summary={null} defaultSectionsOpen />
                </div>
              </Expand>
            )
          })}
        </div>
      </Card>
    </div>
  )
}
