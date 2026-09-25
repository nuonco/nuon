import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'

export type TActivityPinCard = {
  id: string
  kind: 'action' | 'runbook'
  loading?: boolean
  missing?: boolean
  name: string
  onRemove?: () => void
  onRun?: () => void
}

export interface IActivityPins {
  items: TActivityPinCard[]
  panel?: ReactNode
}

const iconFor = (kind: TActivityPinCard['kind']) => (
  <Icon
    variant={kind === 'runbook' ? 'BookIcon' : 'TerminalWindowIcon'}
    size={16}
    className="text-cool-grey-400 shrink-0"
  />
)

export const ActivityPins = ({ items, panel }: IActivityPins) => (
  <div className="flex flex-col gap-3">
    <div className="flex items-center justify-between gap-4">
      {items.length ? (
        <Text variant="body" weight="strong" role="heading" level={2}>
          Pinned
        </Text>
      ) : (
        <Text theme="neutral">
          Pin the runbooks and actions you run often. They stay at the top of
          this page.
        </Text>
      )}
      {panel}
    </div>
    {items.length ? (
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3">
        {items.map((item) => {
          const name = item.missing ? 'Unavailable' : item.name

          if (item.onRun && !item.onRemove && !item.loading) {
            return (
              <Tooltip
                key={item.id}
                className="!w-full flex min-w-0"
                position="top"
                tipContent={`Run ${item.name}`}
              >
                <button
                  type="button"
                  className={cn(
                    'group flex items-center gap-2 w-full min-w-0 p-3 border border-solid rounded-md shadow-sm text-left cursor-pointer',
                    'bg-transparent font-sans text-inherit',
                    'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-700/60',
                    'focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-2 focus-visible:outline-primary-400/80'
                  )}
                  aria-label={`Run ${item.name}`}
                  onClick={item.onRun}
                >
                  {iconFor(item.kind)}
                  <Text
                    variant="body"
                    weight="stronger"
                    className="truncate"
                  >
                    {item.name}
                  </Text>
                  <Icon
                    variant="PlayIcon"
                    size={16}
                    className="ml-auto shrink-0 text-primary-600 dark:text-primary-400 opacity-70 group-hover:opacity-100 group-focus-visible:opacity-100"
                  />
                </button>
              </Tooltip>
            )
          }

          return (
            <Card key={item.id} className="!p-3 !gap-3 min-w-0">
              <div className="flex items-center gap-2 min-w-0">
                {iconFor(item.kind)}
                <Text
                  variant="body"
                  weight="stronger"
                  className="truncate"
                  title={item.loading ? undefined : name}
                  loading={item.loading}
                  loadingWidth={14}
                >
                  {name}
                </Text>
              </div>
              {item.missing ? (
                <Text variant="subtext" theme="neutral">
                  This shortcut is no longer on the install.
                </Text>
              ) : null}
              {item.onRemove ? (
                <Button
                  className="self-start"
                  variant="secondary"
                  onClick={item.onRemove}
                >
                  Unpin
                </Button>
              ) : null}
            </Card>
          )
        })}
      </div>
    ) : null}
  </div>
)
