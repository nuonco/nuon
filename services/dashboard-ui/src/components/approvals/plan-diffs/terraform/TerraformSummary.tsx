import { Text } from '@/components/common/Text'
import type { TTerraformChangeAction } from '@/types'

type TTerraformSummary = Record<TTerraformChangeAction, number>

export const TerraformSummary = ({
  summary,
}: {
  summary: TTerraformSummary
}) => {
  return (
    <div className="px-4 py-3 sm:px-6 border-b bg-cool-grey-100 dark:bg-dark-grey-800">
      <div className="flex space-x-4">
        <div className="flex items-center gap-1.5">
          <Text variant="base" theme="success" weight="strong">
            {summary.create}
          </Text>
          <Text variant="subtext" theme="neutral">
            to create
          </Text>
        </div>
        <div className="flex items-center gap-1.5">
          <Text variant="base" theme="warn" weight="strong">
            {summary.update}
          </Text>
          <Text variant="subtext" theme="neutral">
            to update
          </Text>
        </div>
        <div className="flex items-center gap-1.5">
          <Text variant="base" theme="error" weight="strong">
            {summary.delete}
          </Text>
          <Text variant="subtext" theme="neutral">
            to delete
          </Text>
        </div>
        <div className="flex items-center gap-1.5">
          <Text variant="base" theme="brand" weight="strong">
            {summary.replace}
          </Text>
          <Text variant="subtext" theme="neutral">
            to replace
          </Text>
        </div>

        <div className="flex items-center gap-1.5">
          <Text variant="base" theme="info" weight="strong">
            {summary.read}
          </Text>
          <Text variant="subtext" theme="neutral">
            to read
          </Text>
        </div>

        <div className="flex items-center gap-1.5">
          <Text variant="base" theme="neutral" weight="strong">
            {summary?.noop || summary['no-op']}
          </Text>
          <Text variant="subtext" theme="neutral">
            noop
          </Text>
        </div>
      </div>
    </div>
  )
}
