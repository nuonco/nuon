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
  if (title.toLowerCase().includes('seed')) return 'Database'
  if (title.toLowerCase().includes('reset')) return 'ArrowCounterClockwise'
  if (title.toLowerCase().includes('sync')) return 'ArrowsClockwise'
  if (title.toLowerCase().includes('generate')) return 'MagicWand'
  if (title.toLowerCase().includes('mock')) return 'Cube'
  if (title.toLowerCase().includes('log')) return 'Terminal'
  if (title.toLowerCase().includes('cache')) return 'Lightning'
  if (title.toLowerCase().includes('migrate')) return 'ArrowRight'
  if (title.toLowerCase().includes('test')) return 'Flask'

  switch (variant) {
    case 'warning':
      return 'Warning'
    case 'danger':
      return 'Warning'
    default:
      return 'Play'
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
