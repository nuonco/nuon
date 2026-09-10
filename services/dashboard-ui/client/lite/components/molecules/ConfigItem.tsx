import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Icon, type TIconVariant } from '../atoms/Icon'
import { Text } from '../atoms/Text'
import { ID } from './ID'

export interface IConfigItem
  extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> {
  icon?: TIconVariant
  name?: string
  id?: string
  status?: ReactNode
  metadata?: ReactNode
  loading?: boolean
}

export const ConfigItem = ({
  icon = 'CubeIcon',
  name,
  id,
  status,
  metadata,
  loading = false,
  className,
  ...props
}: IConfigItem) => (
  <button
    type="button"
    className={cn(
      'flex w-full cursor-pointer items-center gap-3 rounded-lg px-2 py-2 text-left transition-colors hover:bg-surface-accent focus-ring',
      className
    )}
    {...props}
  >
    <span className="flex size-7 shrink-0 items-center justify-center rounded-md bg-surface-accent text-secondary">
      <Icon variant={icon} size={16} />
    </span>
    <span className="flex min-w-0 flex-1 flex-col gap-0.5">
      <span className="flex min-w-0 items-center gap-2">
        <Text
          weight="medium"
          className="min-w-0 truncate"
          loading={loading}
          loadingWidth={14}
        >
          {name ?? '\u2014'}
        </Text>
        {status ? (
          <span className="flex shrink-0 items-center gap-1">{status}</span>
        ) : null}
      </span>
      <span className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-0.5 text-tertiary">
        {loading ? (
          <Text variant="caption" loading loadingWidth={10} />
        ) : (
          <>
            {id ? <ID value={id} truncate copyable={false} /> : null}
            {metadata}
          </>
        )}
      </span>
    </span>
    <Icon variant="CaretRightIcon" size={14} className="shrink-0 text-tertiary" />
  </button>
)
