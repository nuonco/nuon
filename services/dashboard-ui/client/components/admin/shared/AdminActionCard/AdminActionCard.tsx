import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'

export interface IAdminActionCard {
  title: string
  description: string
  variant?: 'default' | 'warning' | 'danger'
  isLoading: boolean
  onClick: () => void
}

const getActionIcon = (title: string, variant: IAdminActionCard['variant']) => {
  if (title.toLowerCase().includes('add') || title.toLowerCase().includes('support user')) return 'UserPlus'
  if (title.toLowerCase().includes('remove') || title.toLowerCase().includes('support user')) return 'UserMinus'
  if (title.toLowerCase().includes('reprovision')) return 'ArrowClockwise'
  if (title.toLowerCase().includes('restart')) return 'ArrowCounterClockwise'
  if (title.toLowerCase().includes('teardown') || title.toLowerCase().includes('force')) return 'Trash'
  if (title.toLowerCase().includes('shutdown') && title.toLowerCase().includes('graceful')) return 'Power'
  if (title.toLowerCase().includes('shutdown') || title.toLowerCase().includes('stop')) return 'Stop'
  if (title.toLowerCase().includes('invalidate') || title.toLowerCase().includes('token')) return 'Key'
  if (title.toLowerCase().includes('debug')) return 'Bug'
  if (title.toLowerCase().includes('update') || title.toLowerCase().includes('sandbox')) return 'Upload'

  switch (variant) {
    case 'warning':
      return 'Warning'
    case 'danger':
      return 'Warning'
    default:
      return 'Play'
  }
}

const getVariantStyles = (variant: IAdminActionCard['variant']) => {
  switch (variant) {
    case 'danger':
      return { buttonVariant: 'danger' as const }
    default:
      return { buttonVariant: 'secondary' as const }
  }
}

export const AdminActionCard = ({
  title,
  description,
  variant = 'default',
  isLoading,
  onClick,
}: IAdminActionCard) => {
  const styles = getVariantStyles(variant)

  return (
    <div className="space-y-3 p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors">
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
