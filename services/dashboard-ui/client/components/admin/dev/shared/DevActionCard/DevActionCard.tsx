import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'

export interface IDevActionCard {
  title: string
  description: string
  variant?: 'default' | 'warning' | 'danger'
  isLoading: boolean
  onClick: () => void
}

const getActionIcon = (
  title: string,
  variant: IDevActionCard['variant']
) => {
  if (title.toLowerCase().includes('seed')) return 'DatabaseIcon'
  if (title.toLowerCase().includes('reset')) return 'ArrowCounterClockwiseIcon'
  if (title.toLowerCase().includes('sync')) return 'ArrowsClockwiseIcon'
  if (title.toLowerCase().includes('generate')) return 'MagicWandIcon'
  if (title.toLowerCase().includes('mock')) return 'CubeIcon'
  if (title.toLowerCase().includes('log')) return 'TerminalIcon'
  if (title.toLowerCase().includes('cache')) return 'LightningIcon'
  if (title.toLowerCase().includes('migrate')) return 'ArrowRightIcon'
  if (title.toLowerCase().includes('test')) return 'FlaskIcon'

  switch (variant) {
    case 'warning':
      return 'WarningIcon'
    case 'danger':
      return 'WarningIcon'
    default:
      return 'PlayIcon'
  }
}

const getVariantStyles = (variant: IDevActionCard['variant']) => {
  switch (variant) {
    case 'danger':
      return { buttonVariant: 'danger' as const }
    default:
      return { buttonVariant: 'secondary' as const }
  }
}

export const DevActionCard = ({
  title,
  description,
  variant = 'default',
  isLoading,
  onClick,
}: IDevActionCard) => {
  const styles = getVariantStyles(variant)

  return (
    <div className="space-y-3 p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-teal-300 dark:hover:border-teal-600 transition-colors">
      <div className="flex flex-col">
        <Text variant="base" weight="strong">
          {title}
        </Text>
        <Text variant="subtext" className="text-gray-600 dark:text-gray-300">
          {description}
        </Text>
      </div>

      <Button
        onClick={onClick}
        disabled={isLoading}
        variant={styles.buttonVariant}
      >
        {isLoading ? (
          <>
            <Icon variant="Loading" className="animate-spin" />
            Executing...
          </>
        ) : (
          <>
            <Icon variant={getActionIcon(title, variant)} />
            {title}
          </>
        )}
      </Button>
    </div>
  )
}
