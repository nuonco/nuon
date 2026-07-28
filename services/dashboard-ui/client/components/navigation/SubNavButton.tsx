import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { usePageSidebar } from '@/hooks/use-page-sidebar'
import type { TNavAction } from '@/types'
import { cn } from '@/utils/classnames'

export const SubNavButton = ({
  iconVariant,
  text,
  onClick,
  isActive = false,
}: Omit<TNavAction, 'type' | 'key'>) => {
  const { isPageSidebarOpen } = usePageSidebar()

  const button = (
    <button
      type="button"
      onClick={onClick}
      aria-current={isActive ? 'page' : undefined}
      className={cn(
        'link font-sans transition-colors cursor-pointer text-left',
        'flex items-center gap-4 overflow-hidden rounded-md py-2.5 px-3 w-full',
        'text-[14px] h-[36px] leading-[21px] tracking-[-0.2px]',
        'hover:bg-black/5 hover:dark:bg-white/10',
        'focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-0 focus-visible:outline-primary-400/80',
        'has-[svg]:flex has-[svg]:items-center has-[svg]:gap-1.5',
        isActive
          ? 'text-primary-800 dark:text-primary-400 bg-primary-200 dark:bg-primary-600/25'
          : 'text-cool-grey-800 dark:text-cool-grey-400'
      )}
    >
      <span>
        {iconVariant ? <Icon variant={iconVariant} weight="bold" /> : null}
      </span>
      <span
        className={cn(
          'whitespace-nowrap transition-all duration-fastest ease-cubic md:ml-2 w-fit',
          {
            'md:opacity-100 md:w-full': isPageSidebarOpen,
            'md:opacity-0 w-0': !isPageSidebarOpen,
          }
        )}
      >
        {text}
      </span>
    </button>
  )

  return (
    <Tooltip
      className="w-full"
      position="right"
      tipContent={
        <Text variant="subtext" weight="stronger">
          {text}
        </Text>
      }
      tipContentClassName={cn('hidden', {
        'md:flex w-max': !isPageSidebarOpen,
      })}
    >
      {button}
    </Tooltip>
  )
}
