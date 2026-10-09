import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Expand } from '@/components/common/Expand'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import { type DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { AppConfigDiff } from '@/components/diffs/plan-diff-switch'
import { filterExcludedSections } from '../ConfigStep/lib'
import { STEP_GUTTER, StepBlock, StepRowList } from '../../shared/StepLayout'
import { cn } from '@/utils/classnames'
import { useInstallLink } from '@/hooks/use-install-path'
import {
  getApprovalResponseTheme,
  getApprovalResponseType,
} from '@/utils/approval-utils'

export interface PlanInstallDiff {
  installId: string
  installName: string
  installLabels?: Record<string, string>
  sections: DiffSectionData[]
  summary: { added: number; removed: number; changed: number } | null
  isLoading?: boolean
  versionLabel?: string
  previousSha?: string
  sha?: string
  message?: string
  author?: string
  createdAt?: string
}

const INSTALL_CARD =
  'rounded-xl border bg-white shadow-sm dark:bg-dark-grey-900'

interface IPlanGroupStep {
  installs: PlanInstallDiff[]
  groupName?: string
  labelColors?: Record<string, string>
  orgId?: string
  hasResponse: boolean
  responseType?: string
  showApproveBar: boolean
  isInProgress: boolean
  actions?: ReactNode
  hideHeading?: boolean
  onSelectInstall?: (installId: string) => void
  installFacts?: Record<
    string,
    {
      labels?: Record<string, string>
      region?: string
      status?: string
      detail?: string
      appliedConfigId?: string
    }
  >
}

export const PlanGroupStep = ({
  installs,
  groupName,
  labelColors,
  orgId,
  hasResponse,
  responseType,
  showApproveBar,
  isInProgress: _isInProgress,
  actions,
  hideHeading = false,
  onSelectInstall,
  installFacts,
}: IPlanGroupStep) => {
  const installLink = useInstallLink()
  return (
    <>
      {(hasResponse || showApproveBar) && (
        <StepBlock className={hideHeading ? '!px-0 !pt-0' : undefined}>
          {hasResponse && (
            <Banner theme={getApprovalResponseTheme(responseType)}>
              <Text weight="strong">
                {getApprovalResponseType(responseType)
                  ? `Plan was ${getApprovalResponseType(responseType)}`
                  : 'Plan responded'}
              </Text>
            </Banner>
          )}

          {showApproveBar && (
            <Banner className="@container" theme="neutral">
              <div className="flex flex-col gap-3 @md:flex-row @md:items-center @md:justify-between">
                <div className="flex min-w-0 flex-col">
                  <Text weight="strong">
                    Install group plan requires review
                  </Text>
                  <Text variant="subtext" theme="neutral">
                    Review the changes below, then approve to deploy or skip
                    this install group.
                  </Text>
                </div>
                {actions ? (
                  <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">
                    {actions}
                  </div>
                ) : null}
              </div>
            </Banner>
          )}
        </StepBlock>
      )}

      {hideHeading ? null : (
        <StepBlock>
          <div className="flex items-center gap-3">
            <Icon variant="ListChecksIcon" size="16" />
            <Text variant="base" weight="strong">
              {groupName || 'Install group'}
            </Text>
            <Text variant="subtext" theme="neutral">
              {installs.length} {installs.length === 1 ? 'install' : 'installs'}
            </Text>
          </div>
        </StepBlock>
      )}

      <StepRowList className={hideHeading ? 'gap-2 divide-y-0' : undefined}>
        {installs.map((inst) => {
          const sections = filterExcludedSections(inst.sections)
          const total =
            (inst.summary?.added ?? 0) +
            (inst.summary?.removed ?? 0) +
            (inst.summary?.changed ?? 0)
          const hasChanges = total > 0 && sections.length > 0
          const facts = installFacts?.[inst.installId]
          const labels = inst.installLabels ?? facts?.labels
          const labelEntries = labels ? Object.entries(labels) : []
          const region = facts?.region
          const installLabel = inst.installName || inst.installId

          const changeSummary = inst.isLoading ? (
            <Text variant="subtext" theme="neutral" className="shrink-0">
              Loading…
            </Text>
          ) : (
            <ChangeCountSummary
              added={inst.summary?.added ?? 0}
              updated={inst.summary?.changed ?? 0}
              removed={inst.summary?.removed ?? 0}
              emptyText="No changes"
              className="shrink-0"
            />
          )

          const name =
            orgId && inst.installId ? (
              <Link
                href={installLink({
                  orgId: orgId,
                  installId: inst.installId,
                })}
                textVariant="body"
                className="font-strong break-words"
              >
                {installLabel}
              </Link>
            ) : (
              <Text weight="strong" className="break-words">
                {installLabel}
              </Text>
            )

          const heading = (
            <div className="flex flex-wrap items-center gap-x-3 gap-y-2 min-w-0 flex-1">
              {name}
              {hideHeading ? changeSummary : null}
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
              {region ? (
                <Text variant="subtext" theme="neutral" className="shrink-0">
                  {region}
                </Text>
              ) : null}
              {facts?.status || facts?.detail ? (
                <span className="ml-auto flex shrink-0 items-center gap-2">
                  {facts.status ? <Status status={facts.status} /> : null}
                  {facts.detail ? (
                    <Text variant="subtext" theme="neutral">
                      {facts.detail}
                    </Text>
                  ) : null}
                </span>
              ) : null}
              {onSelectInstall ? (
                <Button
                  size="sm"
                  variant="secondary"
                  className={
                    facts?.status || facts?.detail
                      ? 'shrink-0'
                      : 'ml-auto shrink-0'
                  }
                  onClick={(event) => {
                    event.stopPropagation()
                    onSelectInstall(inst.installId)
                  }}
                >
                  View details
                </Button>
              ) : null}
            </div>
          )

          if (!hasChanges) {
            return (
              <div
                key={inst.installId}
                className={cn(
                  'flex items-start gap-3 py-3',
                  STEP_GUTTER,
                  hideHeading && INSTALL_CARD
                )}
              >
                {heading}
                {hideHeading ? null : changeSummary}
                <Icon
                  variant="CaretDownIcon"
                  className="invisible shrink-0"
                  aria-hidden
                />
              </div>
            )
          }

          return (
            <Expand
              key={inst.installId}
              id={`plan-install-${inst.installId}`}
              interactiveHeading
              toggleLabel={`Show plan changes for ${installLabel}`}
              toggleContent={hideHeading ? undefined : changeSummary}
              heading={heading}
              headerClassName={cn(STEP_GUTTER, 'py-3')}
              className={hideHeading ? INSTALL_CARD : undefined}
            >
              <div className="border-t bg-black/[0.015] dark:bg-white/[0.0075]">
                <AppConfigDiff
                  sections={sections}
                  summary={null}
                  defaultSectionsOpen={false}
                  embedded
                />
              </div>
            </Expand>
          )
        })}
      </StepRowList>
    </>
  )
}
