import {
  forwardRef,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type HTMLAttributes,
  type KeyboardEvent,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Checkbox } from '@/components/common/form/CheckboxInput'

export interface IFilterOption {
  checked: boolean
  isolated: boolean
  onToggle: () => void
  onIsolate: () => void
  label: ReactNode
  textValue: string
  description?: ReactNode
  leading?: ReactNode
  rail?: string
  disabled?: boolean
  tabIndex?: number
  onKeyDown?: (event: KeyboardEvent<HTMLDivElement>) => void
}

export const FilterOption = forwardRef<HTMLDivElement, IFilterOption>(
  (
    {
      checked,
      isolated,
      onToggle,
      onIsolate,
      label,
      textValue,
      description,
      leading,
      rail,
      disabled = false,
      tabIndex = -1,
      onKeyDown,
    },
    ref
  ) => (
    <div
      ref={ref}
      role="group"
      tabIndex={disabled ? -1 : tabIndex}
      aria-label={`${textValue}, ${checked ? 'included' : 'excluded'}. Space toggles. Enter ${
        isolated ? 'resets filters' : 'shows only this option'
      }.`}
      aria-disabled={disabled || undefined}
      data-filter-option
      onClick={(event) => {
        if (disabled || (event.target as HTMLElement).closest('input')) return
        onIsolate()
      }}
      onKeyDown={(event) => {
        onKeyDown?.(event)
        if (event.defaultPrevented || disabled) return
        if (event.key === ' ') {
          event.preventDefault()
          event.stopPropagation()
          onToggle()
        }
        if (event.key === 'Enter') {
          event.preventDefault()
          event.stopPropagation()
          onIsolate()
        }
      }}
      className={cn(
        'group/filter relative flex min-h-8 cursor-pointer items-stretch rounded-md outline-none transition-colors',
        'hover:bg-cool-grey-500/8 focus-visible:bg-cool-grey-500/8',
        'focus-visible:outline-1 focus-visible:-outline-offset-2 focus-visible:outline-primary-400/80',
        disabled && 'cursor-not-allowed opacity-50'
      )}
    >
      {rail ? (
        <span aria-hidden className={cn('w-1 shrink-0 rounded-sm', rail)} />
      ) : null}
      <span className="flex min-w-0 flex-1 items-center gap-2 px-2 py-1.5">
        <Checkbox
          checked={checked}
          disabled={disabled}
          tabIndex={-1}
          aria-label={`Include ${textValue}`}
          onChange={onToggle}
        />
        {leading ? (
          <span className="flex shrink-0 items-center">{leading}</span>
        ) : null}
        <span className="flex min-w-0 flex-1 flex-col items-start">
          {typeof label === 'string' ? (
            <Text variant="subtext" className="truncate">
              {label}
            </Text>
          ) : (
            label
          )}
          {description ? (
            <Text variant="subtext" theme="neutral" className="truncate">
              {description}
            </Text>
          ) : null}
        </span>
        <Text
          variant="subtext"
          theme="neutral"
          aria-hidden
          className="ml-2 shrink-0 opacity-0 transition-opacity group-hover/filter:opacity-100 group-focus-visible/filter:opacity-100"
        >
          {isolated ? 'Reset' : 'Only'}
        </Text>
      </span>
    </div>
  )
)

FilterOption.displayName = 'FilterOption'

export interface IFilterMenuOption<T extends string> {
  value: T
  label: ReactNode
  textValue?: string
  description?: ReactNode
  leading?: ReactNode
  rail?: string
  disabled?: boolean
}

export interface IFilterMenu<T extends string>
  extends Omit<HTMLAttributes<HTMLDivElement>, 'onToggle'> {
  label: string
  options: readonly IFilterMenuOption<T>[]
  selected: Set<T>
  onToggle: (value: T) => void
  onIsolate: (value: T) => void
  onReset: () => void
}

export const FilterMenu = <T extends string>({
  label,
  options,
  selected,
  onToggle,
  onIsolate,
  onReset,
  className,
  ...props
}: IFilterMenu<T>) => {
  const refs = useRef<Array<HTMLDivElement | null>>([])
  const [activeIndex, setActiveIndex] = useState(0)

  const enabledIndices = useMemo(
    () => options.flatMap((option, index) => (option.disabled ? [] : [index])),
    [options]
  )

  const focusAt = useCallback((index: number) => {
    setActiveIndex(index)
    refs.current[index]?.focus()
  }, [])

  useEffect(() => {
    if (!enabledIndices.includes(activeIndex)) {
      setActiveIndex(enabledIndices[0] ?? -1)
    }
  }, [activeIndex, enabledIndices])

  const move = (
    current: number,
    direction: 1 | -1,
    event: KeyboardEvent<HTMLDivElement>
  ) => {
    event.preventDefault()
    event.stopPropagation()
    const position = enabledIndices.indexOf(current)
    const next =
      position < 0
        ? direction === 1
          ? enabledIndices[0]
          : enabledIndices.at(-1)
        : enabledIndices[
            (position + direction + enabledIndices.length) %
              enabledIndices.length
          ]
    if (next !== undefined) focusAt(next)
  }

  return (
    <div
      role="group"
      aria-label={label}
      className={cn('flex min-w-56 flex-col p-1', className)}
      {...props}
    >
      {options.map((option, index) => (
        <FilterOption
          key={option.value}
          ref={(element) => {
            refs.current[index] = element
          }}
          checked={selected.has(option.value)}
          isolated={selected.size === 1 && selected.has(option.value)}
          label={option.label}
          textValue={
            option.textValue ??
            (typeof option.label === 'string' ? option.label : option.value)
          }
          description={option.description}
          leading={option.leading}
          rail={option.rail}
          disabled={option.disabled}
          tabIndex={index === activeIndex ? 0 : -1}
          onToggle={() => onToggle(option.value)}
          onIsolate={() => onIsolate(option.value)}
          onKeyDown={(event) => {
            if (event.key === 'ArrowDown') move(index, 1, event)
            if (event.key === 'ArrowUp') move(index, -1, event)
            if (event.key === 'Home') {
              event.preventDefault()
              const first = enabledIndices[0]
              if (first !== undefined) focusAt(first)
            }
            if (event.key === 'End') {
              event.preventDefault()
              const last = enabledIndices.at(-1)
              if (last !== undefined) focusAt(last)
            }
          }}
        />
      ))}
      <hr className="my-1" />
      <Button
        size="sm"
        variant="ghost"
        className="w-full justify-start"
        onClick={onReset}
      >
        Reset
      </Button>
    </div>
  )
}

export interface IFilterDropdown<T extends string> {
  label: string
  options: readonly IFilterMenuOption<T>[]
  selected: Set<T>
  onToggle: (value: T) => void
  onIsolate: (value: T) => void
  onReset: () => void
  constrained?: boolean
  className?: string
}

export const FilterDropdown = <T extends string>({
  label,
  options,
  selected,
  onToggle,
  onIsolate,
  onReset,
  constrained = selected.size !== options.length,
  className,
}: IFilterDropdown<T>) => (
  <Dropdown
    id={`filter-${label.replace(/\s+/g, '-').toLowerCase()}`}
    alignment="right"
    variant="ghost"
    size="sm"
    buttonClassName={className}
    icon={<Icon variant="FunnelIcon" size={14} />}
    iconAlignment="left"
    buttonText={`${label}${constrained ? ` (${selected.size})` : ''}`}
  >
    <FilterMenu
      label={`${label} filters`}
      options={options}
      selected={selected}
      onToggle={onToggle}
      onIsolate={onIsolate}
      onReset={onReset}
    />
  </Dropdown>
)
