import { useState, useCallback } from 'react'

interface ISearchInput {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
}

export const SearchInput = ({ value, onChange, placeholder = 'Search...', className = '' }: ISearchInput) => {
  const [local, setLocal] = useState(value)

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        onChange(local)
      }
    },
    [local, onChange],
  )

  return (
    <input
      type="text"
      value={local}
      onChange={(e) => setLocal(e.target.value)}
      onBlur={() => onChange(local)}
      onKeyDown={handleKeyDown}
      placeholder={placeholder}
      className={`block w-full rounded-md border-0 py-1.5 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-primary-600 sm:text-sm sm:leading-6 ${className}`}
    />
  )
}
