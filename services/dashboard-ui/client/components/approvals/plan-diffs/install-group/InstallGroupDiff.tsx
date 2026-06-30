import { Badge, type TBadgeTheme } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponentType } from '@/types'
import { ChangeCountSummary } from '../ChangeCountSummary'
import { useConfigDiffFocus } from '../config-diff-focus'
import type { DiffSectionData } from '../app-config/AppConfigDiff'

export type InstallDiffEntry = {
  installId: string
  installName: string
  installLabels?: Record<string, string>
  status?: string
  sections: DiffSectionData[]
  summary: { added: number; removed: number; changed: number }
  sandboxChanged?: boolean
  stackChanged?: boolean
}

export interface IInstallGroupDiff {
  groupName: string
  installs: InstallDiffEntry[]
  isLoading?: boolean
}

const SKELETON_ROW_WIDTHS = ['9rem', '7rem', '11rem', '8rem']

const GroupHeader = ({ groupName, children }: { groupName: string; children?: React.ReactNode }) => (
  <div className="px-4 sm:px-6 py-4 border-b">
    <div className="flex items-center gap-3">
      <Icon variant="ListChecksIcon" size="16" />
      <Text variant="base" weight="strong">{groupName}</Text>
      {children}
    </div>
  </div>
)

const InstallGroupDiffSkeleton = ({ groupName, rows = 3 }: { groupName: string; rows?: number }) => (
  <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
    <GroupHeader groupName={groupName}>
      <Skeleton width="4rem" height="0.875rem" />
    </GroupHeader>
    <div className="flex flex-col divide-y">
      {Array.from({ length: rows }).map((_, idx) => (
        <div key={idx} className="flex items-center gap-3 px-4 py-3">
          <Skeleton width={SKELETON_ROW_WIDTHS[idx % SKELETON_ROW_WIDTHS.length]} height="0.875rem" />
          <div className="flex items-center gap-2 ml-auto">
            <Skeleton width="1.5rem" height="0.75rem" />
            <Skeleton width="1.5rem" height="0.75rem" />
            <Skeleton width="1.5rem" height="0.75rem" />
          </div>
        </div>
      ))}
    </div>
  </Card>
)

const OP_VERB: Record<string, string> = { add: 'Add', change: 'Update', remove: 'Remove' }
const OP_THEME: Record<string, TBadgeTheme> = { add: 'success', change: 'warn', remove: 'error' }

type ImpactItem = { name: string; op: string; componentType?: string; isComponent: boolean }

const flattenImpact = (sections: DiffSectionData[]): ImpactItem[] =>
  sections.flatMap((section) =>
    section.entities.map((entity) => ({
      name: entity.name,
      op: entity.op,
      componentType: entity.componentType,
      isComponent: section.sectionKey === 'components',
    }))
  )

const ChangeSummary = ({ install }: { install: InstallDiffEntry }) => {
  const { added, changed, removed } = install.summary
  const updated = changed + (install.sandboxChanged ? 1 : 0) + (install.stackChanged ? 1 : 0)
  return (
    <ChangeCountSummary
      added={added}
      updated={updated}
      removed={removed}
      className="ml-auto shrink-0"
    />
  )
}

export const InstallGroupDiff = ({ groupName, installs, isLoading = false }: IInstallGroupDiff) => {
  const focusCtx = useConfigDiffFocus()

  if (isLoading && installs.length === 0) {
    return <InstallGroupDiffSkeleton groupName={groupName} />
  }

  if (installs.length === 0) {
    return (
      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 py-3 text-center">
          <EmptyState
            emptyTitle="No install changes"
            emptyMessage="No installs in this group will be affected by this plan."
            variant="diagram"
            size="sm"
          />
        </div>
      </Card>
    )
  }

  return (
    <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
      <GroupHeader groupName={groupName}>
        <Text variant="subtext" theme="neutral">
          {installs.length} {installs.length === 1 ? 'install' : 'installs'}
        </Text>
      </GroupHeader>

      <div className="flex flex-col divide-y">
        {installs.map((install) => {
          const totalChanges = install.summary.added + install.summary.removed + install.summary.changed
          const hasChanges = totalChanges > 0 || !!install.sandboxChanged || !!install.stackChanged
          const labelEntries = install.installLabels ? Object.entries(install.installLabels) : []

          const heading = (
            <div className="flex items-center gap-3 w-full">
              <Text weight="strong">{install.installName || install.installId}</Text>
              {labelEntries.map(([k, v]) => (
                <LabelBadge key={k} labelKey={k} labelValue={v} size="sm" variant="code" className="shrink-0" />
              ))}
              <ChangeSummary install={install} />
            </div>
          )

          if (!hasChanges || install.sections.length === 0) {
            return (
              <div key={install.installId} className="flex items-center gap-2 px-4 py-3">
                {heading}
                <Icon variant="CaretDownIcon" className="invisible shrink-0" aria-hidden />
              </div>
            )
          }

          return (
            <Expand
              key={install.installId}
              id={`install-group-diff-${install.installId}`}
              heading={heading}
              headerClassName="px-4 py-3"
            >
              <div className="px-4 py-4 flex flex-col gap-3 border-t border-cool-grey-100 dark:border-dark-grey-800 bg-black/[0.015] dark:bg-white/[0.0075]">
                <Text variant="subtext" theme="neutral">
                  The following will be redeployed to {install.installName || install.installId}:
                </Text>
                <div className="flex flex-col divide-y divide-cool-grey-100 dark:divide-dark-grey-800">
                  {flattenImpact(install.sections).map((item, idx) => {
                    const sectionKey = item.isComponent
                      ? 'components'
                      : item.name === 'Stack'
                        ? 'stack'
                        : 'sandbox'

                    return (
                      <div
                        key={`${item.name}-${idx}`}
                        className="group flex items-center justify-between gap-3 py-2.5 first:pt-0 last:pb-0"
                      >
                        <div className="flex items-center gap-2 min-w-0">
                          {item.isComponent && item.componentType ? (
                            <ComponentType
                              type={item.componentType as TComponentType}
                              colorVariant="color"
                              displayVariant="icon-only"
                              iconSize="16"
                            />
                          ) : (
                            <Icon
                              variant={item.name === 'Stack' ? 'StackIcon' : 'ShippingContainerIcon'}
                              size="16"
                              theme={OP_THEME[item.op] ?? 'neutral'}
                            />
                          )}
                          <Text weight="strong" nowrap className="truncate">
                            {item.name}
                          </Text>
                          {item.isComponent && item.componentType && (
                            <Text variant="subtext" theme="neutral">
                              {item.componentType.replace(/_/g, ' ')}
                            </Text>
                          )}
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          {focusCtx && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="!text-primary-600 dark:!text-primary-500 opacity-0 group-hover:opacity-100 focus-visible:opacity-100 transition-opacity"
                              onClick={() =>
                                focusCtx.requestFocus(
                                  sectionKey,
                                  item.isComponent ? `component.${item.name}` : undefined
                                )
                              }
                            >
                              View details <Icon variant="CaretRightIcon" />
                            </Button>
                          )}
                          <Badge theme={OP_THEME[item.op] ?? 'neutral'} size="sm">
                            {OP_VERB[item.op] ?? item.op}
                          </Badge>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            </Expand>
          )
        })}
      </div>
    </Card>
  )
}
