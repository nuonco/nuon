import { Text } from '@/components/common/Text'

export const DiffEmptyState = () => (
  <div className="rounded-lg bg-cool-grey-100 dark:bg-dark-grey-800 px-4 py-8 text-center">
    <Text as="p" variant="body" weight="strong">
      No changes to show
    </Text>
    <Text as="p" variant="subtext" theme="neutral">
      Clear the search or reset the filters.
    </Text>
  </div>
)
