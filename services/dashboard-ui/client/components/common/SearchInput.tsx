import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { Button } from './Button'
import { Icon } from './Icon'

export interface ISearchInput
  extends Omit<
    HTMLAttributes<HTMLInputElement>,
    'autoComplete' | 'onChange' | 'type'
  > {
  labelClassName?: string
  placeholder?: string
  onChange: (val: string) => void
  onClear?: () => void
  value: string
}

export const SearchInput = ({
  className,
  labelClassName,
  placeholder,
  onChange,
  onClear,
  value,
  ...props
}: ISearchInput) => {
  return (
    <label className={cn('relative w-fit flex', labelClassName)}>
      <Icon
        variant="MagnifyingGlass"
        className="text-cool-grey-500 dark:text-cool-grey-700 absolute top-2.5 left-2"
      />
      <input
        className={cn(
          'rounded-md pl-8 pr-3.5 py-1.5 h-[36px] font-sans md:min-w-80 border text-sm',
          'bg-white dark:bg-dark-grey-900 placeholder:text-cool-grey-500 dark:placeholder:text-cool-grey-700',
          className
        )}
        type="text"
        placeholder={placeholder}
        autoComplete="off"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        {...props}
      />
      {value ? (
        <Button
          className="!p-0.5 !h-fit absolute top-1/2 right-1.5 -translate-y-1/2"
          variant="ghost"
          title="clear search"
          onClick={() => (onClear ? onClear() : onChange(''))}
        >
          <Icon variant="XCircle" />
        </Button>
      ) : null}
    </label>
  )
}
