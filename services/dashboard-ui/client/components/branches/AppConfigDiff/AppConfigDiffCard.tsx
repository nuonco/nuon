import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import type { TConfigDiffFocus } from '@/components/approvals/plan-diffs/config-diff-focus'
import {
  AppConfigDiff,
  type DiffSectionData,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { cn } from '@/utils/classnames'

export interface IAppConfigDiffCard {
  sections: DiffSectionData[]
  summary: { added: number; removed: number; changed: number } | null
  versionLabel?: string | null
  isLoading?: boolean
  isOpen?: boolean
  focus?: TConfigDiffFocus | null
  className?: string
}

export const AppConfigDiffCard = ({
  sections,
  summary,
  versionLabel,
  isLoading = false,
  isOpen = true,
  focus,
  className,
}: IAppConfigDiffCard) => {
  return (
    <Expand
      id="config-changes"
      isOpen={isOpen}
      className={cn(
        'border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden',
        className
      )}
      headerClassName="px-5 py-4"
      heading={
        <div className="flex items-center gap-3 w-full">
          <Text variant="h3" weight="strong">
            Config changes
          </Text>
          {versionLabel && (
            <Text variant="subtext" theme="neutral" family="mono">
              {versionLabel}
            </Text>
          )}
          {summary && (
            <ChangeCountSummary
              added={summary.added}
              updated={summary.changed}
              removed={summary.removed}
              className="ml-auto"
            />
          )}
        </div>
      }
    >
      <div className="p-5 border-t max-h-[70vh] overflow-y-auto">
        <AppConfigDiff
          sections={sections}
          summary={null}
          isLoading={isLoading}
          defaultSectionsOpen={false}
          focus={focus}
        />
      </div>
    </Expand>
  )
}
