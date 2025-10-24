import { Text } from '@/components/common/Text'
import type { THelmPlanSummary } from '@/types'

export const HelmDiffSummary = ({ summary }: { summary: THelmPlanSummary }) => {
  return (
    <div className="px-4 py-3 sm:px-6 border-b bg-cool-grey-100 dark:bg-dark-grey-800">
      <div className="flex space-x-4">
        <div className="flex items-center gap-1.5">
          <Text
            variant="base"
            className="text-green-600 dark:text-green-40"
            weight="strong"
          >
            {summary.add}
          </Text>
          <Text variant="subtext" theme="neutral">
            to add
          </Text>
        </div>
        <div className="flex items-center gap-1.5">
          <Text
            variant="base"
            className="text-orange-600 dark:text-orange-400"
            weight="strong"
          >
            {summary.change}
          </Text>
          <Text variant="subtext" theme="neutral">
            to change
          </Text>
        </div>
        <div className="flex items-center gap-1.5">
          <Text
            variant="base"
            className="text-red-600 dark:text-red-400"
            weight="strong"
          >
            {summary.destroy}
          </Text>
          <Text variant="subtext" theme="neutral">
            to destroy
          </Text>
        </div>
      </div>
    </div>
  )
}
