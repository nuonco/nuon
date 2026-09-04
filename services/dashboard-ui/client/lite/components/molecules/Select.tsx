import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { Dropdown, useDropdown } from '../atoms/Dropdown'
import { Icon } from '../atoms/Icon'
import {
  FIELD_CONTROL_CLASSES,
  FIELD_SIZE_CLASSES,
  type TFieldSize,
} from '../atoms/Input'
import { Text } from '../atoms/Text'
import { SearchInput } from './SearchInput'

export interface ISelectOption {
  value: string
  label: ReactNode
  textValue?: string
  description?: ReactNode
  disabled?: boolean
}

export interface ISelect {
  value?: string
  defaultValue?: string
  onChange?: (value: string) => void
  options: ISelectOption[]
  placeholder?: string
  searchable?: boolean
  searchPlaceholder?: string
  emptyMessage?: string
  size?: TFieldSize
  disabled?: boolean
  loading?: boolean
  onBlur?: () => void
  name?: string
  id?: string
  'aria-invalid'?: boolean
  'aria-describedby'?: string
  className?: string
}

interface IOptionList {
  options: ISelectOption[]
  value?: string
  searchable: boolean
  searchPlaceholder: string
  emptyMessage: string
  onSelect: (value: string) => void
}

const optionText = (option: ISelectOption) =>
  option.textValue ??
  (typeof option.label === 'string' ? option.label : option.value)

const OptionList = ({
  options,
  value,
  searchable,
  searchPlaceholder,
  emptyMessage,
  onSelect,
}: IOptionList) => {
  const dropdown = useDropdown()
  const [query, setQuery] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const optionRefs = useRef<Array<HTMLDivElement | null>>([])
  const searchRef = useRef<HTMLInputElement>(null)
  const typeahead = useRef({ value: '', timer: 0 })

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase()
    return needle
      ? options.filter((option) =>
          optionText(option).toLowerCase().includes(needle)
        )
      : options
  }, [options, query])

  const enabledIndices = useMemo(
    () => filtered.flatMap((option, index) => (option.disabled ? [] : [index])),
    [filtered]
  )

  const focusAt = useCallback((index: number) => {
    setActiveIndex(index)
    optionRefs.current[index]?.focus()
  }, [])

  useEffect(() => {
    const selectedIndex = filtered.findIndex(
      (option) => option.value === value && !option.disabled
    )
    setActiveIndex(
      selectedIndex >= 0 ? selectedIndex : (enabledIndices[0] ?? -1)
    )
  }, [enabledIndices, filtered, value])

  const focusRelative = (current: number, direction: 1 | -1) => {
    const at = enabledIndices.indexOf(current)
    const next =
      at < 0
        ? direction === 1
          ? enabledIndices[0]
          : enabledIndices.at(-1)
        : enabledIndices[
            (at + direction + enabledIndices.length) % enabledIndices.length
          ]
    if (next !== undefined) focusAt(next)
  }

  useEffect(() => {
    dropdown?.registerFocusFirst(() => {
      if (searchable) searchRef.current?.focus()
      else if (enabledIndices[0] !== undefined) focusAt(enabledIndices[0])
    })
    dropdown?.registerFocusLast(() => {
      const last = enabledIndices.at(-1)
      if (last !== undefined) focusAt(last)
    })
    return () => {
      dropdown?.registerFocusFirst(null)
      dropdown?.registerFocusLast(null)
      window.clearTimeout(typeahead.current.timer)
    }
  }, [dropdown, enabledIndices, focusAt, searchable])

  const select = (option: ISelectOption) => {
    if (option.disabled) return
    onSelect(option.value)
    dropdown?.close()
  }

  const onOptionKeyDown = (
    event: KeyboardEvent<HTMLDivElement>,
    option: ISelectOption,
    index: number
  ) => {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault()
      event.stopPropagation()
      focusRelative(index, event.key === 'ArrowDown' ? 1 : -1)
      return
    }
    if (event.key === 'Home' || event.key === 'End') {
      event.preventDefault()
      const next =
        event.key === 'Home' ? enabledIndices[0] : enabledIndices.at(-1)
      if (next !== undefined) focusAt(next)
      return
    }
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      select(option)
      return
    }
    if (searchable || event.key.length !== 1) return
    window.clearTimeout(typeahead.current.timer)
    typeahead.current.value =
      `${typeahead.current.value}${event.key}`.toLowerCase()
    const match = filtered.findIndex(
      (candidate) =>
        !candidate.disabled &&
        optionText(candidate).toLowerCase().startsWith(typeahead.current.value)
    )
    if (match >= 0) focusAt(match)
    typeahead.current.timer = window.setTimeout(() => {
      typeahead.current.value = ''
    }, 500)
  }

  return (
    <div className="flex min-w-56 flex-col gap-1 p-1">
      {searchable ? (
        <SearchInput
          ref={searchRef}
          size="sm"
          value={query}
          placeholder={searchPlaceholder}
          aria-label={searchPlaceholder}
          onValueChange={setQuery}
          onKeyDown={(event) => {
            if (event.key !== 'ArrowDown') return
            event.preventDefault()
            if (enabledIndices[0] !== undefined) focusAt(enabledIndices[0])
          }}
          className="mb-1"
        />
      ) : null}
      <div role="listbox" aria-label="Options" className="flex flex-col">
        {filtered.length ? (
          filtered.map((option, index) => (
            <div
              key={option.value}
              ref={(element) => {
                optionRefs.current[index] = element
              }}
              role="option"
              tabIndex={option.disabled || index !== activeIndex ? -1 : 0}
              aria-selected={option.value === value}
              aria-disabled={option.disabled || undefined}
              onClick={() => select(option)}
              onKeyDown={(event) => onOptionKeyDown(event, option, index)}
              className={cn(
                'flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 outline-none',
                'hover:bg-field-bg-hover focus-visible:bg-field-bg-hover focus-visible:outline-2 focus-visible:outline-focus-ring',
                option.value === value && 'bg-field-option-selected',
                option.disabled && 'cursor-not-allowed opacity-50'
              )}
            >
              <span className="flex min-w-0 flex-1 flex-col">
                <Text variant="caption" color="secondary">
                  {option.label}
                </Text>
                {option.description ? (
                  <Text variant="caption" color="tertiary">
                    {option.description}
                  </Text>
                ) : null}
              </span>
              {option.value === value ? (
                <Icon
                  variant="CheckIcon"
                  size={14}
                  className="shrink-0 text-accent"
                />
              ) : null}
            </div>
          ))
        ) : (
          <Text
            variant="caption"
            color="tertiary"
            className="px-2 py-3 text-center"
          >
            {emptyMessage}
          </Text>
        )}
      </div>
    </div>
  )
}

export const Select = ({
  value: controlledValue,
  defaultValue,
  onChange,
  options,
  placeholder = 'Select an option',
  searchable = false,
  searchPlaceholder = 'Search options',
  emptyMessage = 'No options found',
  size = 'md',
  disabled = false,
  loading = false,
  onBlur,
  name,
  id,
  className,
  'aria-invalid': invalid,
  'aria-describedby': describedBy,
}: ISelect) => {
  const [uncontrolledValue, setUncontrolledValue] = useState(defaultValue)
  const [open, setOpen] = useState(false)
  const value =
    controlledValue !== undefined ? controlledValue : uncontrolledValue
  const selected = options.find((option) => option.value === value)

  const select = (next: string) => {
    if (controlledValue === undefined) setUncontrolledValue(next)
    onChange?.(next)
  }

  return (
    <>
      {name ? <input type="hidden" name={name} value={value ?? ''} /> : null}
      <Dropdown
        open={open}
        onOpenChange={(next) => {
          setOpen(next)
          if (!next) onBlur?.()
        }}
        haspopup="listbox"
        matchTriggerWidth
        stretch
        contentClassName="bg-popover-bg text-popover-text shadow-[var(--popover-shadow)]"
        trigger={
          <button
            id={id}
            type="button"
            disabled={disabled || loading}
            aria-invalid={invalid}
            aria-describedby={describedBy}
            aria-busy={loading || undefined}
            className={cn(
              FIELD_CONTROL_CLASSES,
              FIELD_SIZE_CLASSES[size],
              'flex items-center gap-2 text-left',
              loading && 'skeleton text-transparent [&>*]:invisible',
              className
            )}
          >
            <span
              className={cn(
                'min-w-0 flex-1 truncate',
                !selected && 'text-field-placeholder'
              )}
            >
              {selected?.label ?? placeholder}
            </span>
            <Icon
              variant="CaretDownIcon"
              size={14}
              className="shrink-0 text-tertiary"
            />
          </button>
        }
      >
        <OptionList
          options={options}
          value={value}
          searchable={searchable}
          searchPlaceholder={searchPlaceholder}
          emptyMessage={emptyMessage}
          onSelect={select}
        />
      </Dropdown>
    </>
  )
}
