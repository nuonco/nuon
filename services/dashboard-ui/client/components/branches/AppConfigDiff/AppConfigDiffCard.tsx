import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import type { TConfigDiffFocus } from '@/components/approvals/plan-diffs/config-diff-focus'
import {
  AppConfigDiff,
  type DiffSectionData,
  type TAppConfigDiffPresentation,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { cn } from '@/utils/classnames'
import type { ReactNode } from 'react'

export interface IAppConfigDiffCard {
  sections: DiffSectionData[]
  summary: { added: number; removed: number; changed: number } | null
  versionLabel?: string | null
  title?: string
  headerAction?: ReactNode
  isLoading?: boolean
  isOpen?: boolean
  focus?: TConfigDiffFocus | null
  className?: string
  presentation?: TAppConfigDiffPresentation
  expandId?: string
}

export const AppConfigDiffCard = ({
  sections,
  summary,
  versionLabel,
  title = 'Config changes',
  headerAction,
  isLoading = false,
  isOpen = true,
  focus,
  className,
  presentation = 'diff',
  expandId = 'config-changes',
}: IAppConfigDiffCard) => {
  return (
    <Expand
      id={expandId}
      isOpen={isOpen}
      className={cn(
        'border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden',
        className
      )}
      headerClassName="px-5 py-4"
      heading={
        <div className="flex items-center gap-3 w-full">
          <Text variant="h3" weight="strong">
            {title}
          </Text>
          {versionLabel && (
            <Text variant="subtext" theme="neutral" family="mono">
              {versionLabel}
            </Text>
          )}
          <div className="ml-auto flex items-center gap-3">
            {!isLoading && presentation === 'diff' && (
              <ChangeCountSummary
                added={summary?.added ?? 0}
                updated={summary?.changed ?? 0}
                removed={summary?.removed ?? 0}
                emptyText="No changes"
              />
            )}
            {headerAction}
          </div>
        </div>
      }
    >
      <div className="p-5 border-t max-h-[70vh] overflow-y-auto">
        <AppConfigDiff
          sections={sections}
          summary={null}
          isLoading={isLoading}
          defaultSectionsOpen={presentation === 'snapshot'}
          focus={focus}
          presentation={presentation}
        />
      </div>
    </Expand>
  )
}
