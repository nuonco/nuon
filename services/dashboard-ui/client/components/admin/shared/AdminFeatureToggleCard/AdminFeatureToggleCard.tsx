import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'

interface IAdminFeatureToggleCard {
  onOpenPanel: () => void
}

export const AdminFeatureToggleCard = ({ onOpenPanel }: IAdminFeatureToggleCard) => {
  return (
    <div className="space-y-3 p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors">
      <div className="flex flex-col">
        <Text variant="base" weight="strong">
          Manage organization features
        </Text>
        <Text variant="subtext" className="text-gray-600 dark:text-gray-300">
          Configure feature flags for this organization
        </Text>
      </div>

      <Button onClick={onOpenPanel} variant="secondary">
        <Icon variant="SlidersIcon" />
        Manage features
      </Button>
    </div>
  )
}
