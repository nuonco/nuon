import { forwardRef, useEffect, useState, type ComponentProps } from 'react'
import { SearchInput } from './SearchInput'

export interface IListSearch
  extends Omit<ComponentProps<typeof SearchInput>, 'value' | 'onValueChange'> {
  value: string
  onValueChange: (value: string) => void
  debounceMs?: number
}

export const ListSearch = forwardRef<HTMLInputElement, IListSearch>(
  ({ value, onValueChange, debounceMs = 300, ...props }, ref) => {
    const [draft, setDraft] = useState(value)

    useEffect(() => {
      setDraft(value)
    }, [value])

    useEffect(() => {
      if (draft === value) return

      const timeout = window.setTimeout(() => onValueChange(draft), debounceMs)
      return () => window.clearTimeout(timeout)
    }, [debounceMs, draft, onValueChange, value])

    return (
      <SearchInput
        ref={ref}
        value={draft}
        onValueChange={setDraft}
        {...props}
      />
    )
  }
)

ListSearch.displayName = 'ListSearch'
