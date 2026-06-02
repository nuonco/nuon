import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'

interface IAdminEnvAccentsCard {
  onOpenPanel: () => void
}

export const AdminEnvAccentsCard = ({ onOpenPanel }: IAdminEnvAccentsCard) => {
  return (
    <div className="space-y-3 p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors">
      <div className="flex flex-col">
        <Text variant="base" weight="strong">
          Manage environment accents
        </Text>
        <Text variant="subtext" className="text-gray-600 dark:text-gray-300">
          Map install labels (e.g. env=production) to accent colors so
          prod/non-prod installs are easy to tell apart.
        </Text>
      </div>

      <Button onClick={onOpenPanel} variant="secondary">
        <Icon variant="PaletteIcon" />
        Manage env accents
      </Button>
    </div>
  )
}
