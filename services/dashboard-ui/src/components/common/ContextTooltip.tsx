import type { ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Button } from './Button'
import { Icon } from './Icon'
import { Text } from './Text'
import { Tooltip, type ITooltip } from './Tooltip'

export type TContextTooltipItem = {
  id: string
  href?: string
  leftContent?: ReactNode
  title: string
  subtitle?: string | ReactNode
  rightContent?: ReactNode
  onClick?: () => void
}

interface IContextTooltip
  extends Omit<ITooltip, 'tipContent' | 'tipContentClassName'> {
  items: TContextTooltipItem[]
  title?: string
  showCount?: boolean
  maxHeight?: string
  width?: string
  onItemClick?: (item: TContextTooltipItem) => void
}

export const ContextTooltip = ({
  children,
  items,
  title,
  showCount = false,
  maxHeight = 'max-h-40',
  width = 'w-52',
  position = 'bottom',
  onItemClick,
  ...props
}: IContextTooltip) => {
  const handleItemClick = (item: TContextTooltipItem) => {
    item.onClick?.()
    onItemClick?.(item)
  }

  return (
    <Tooltip
      position={position}
      tipContentClassName="!p-0"
      tipContent={
        <div className={`${width}`}>
          {title ? (
            <Text
              className="px-3 py-2 border-b flex items-center justify-between !leading-none"
              variant="subtext"
              weight="strong"
            >
              {title}
              {showCount && <span>{items.length}</span>}
            </Text>
          ) : null}
          <div className={`divide-y ${maxHeight} overflow-y-auto`}>
            {items.map((item) => {
              const itemContent = (
                <>
                  {item.leftContent ? (
                    <span className="h-full">{item.leftContent}</span>
                  ) : null}
                  <span className="flex flex-col text-left">
                    <Text
                      className="max-w-36 truncate"
                      variant="label"
                      weight="strong"
                    >
                      {item.title}
                    </Text>
                    {item.subtitle ? (
                      typeof item.subtitle === 'string' ? (
                        <Text variant="label" theme="neutral">
                          {item.subtitle}
                        </Text>
                      ) : (
                        item.subtitle
                      )
                    ) : null}
                  </span>
                </>
              )
              return (
                <div
                  key={item.id}
                  className={cn(
                    `last-of-type:rounded-b-lg overflow-hidden shrink-0 grow-0 ${width}`,
                    {
                      'first-of-type:rounded-t-lg': !title,
                    }
                  )}
                >
                  {!item.href && !item.onClick ? (
                    <span className="grid grid-cols-[auto_1fr] justify-between gap-2 w-full px-3 py-2">
                      {itemContent}
                    </span>
                  ) : (
                    <Button
                      className="!flex w-full !rounded-none !h-[unset] !px-3 !py-2"
                      variant="ghost"
                      href={item.href}
                      onClick={
                        item.href ? undefined : () => handleItemClick(item)
                      }
                    >
                      <span className="grid grid-cols-[auto_1fr_auto] justify-between gap-2 w-full">
                        {itemContent}
                        {item.rightContent || (
                          <Icon
                            className="self-center ml-auto"
                            variant="CaretRight"
                          />
                        )}
                      </span>
                    </Button>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      }
      {...props}
    >
      {children}
    </Tooltip>
  )
}
