'use client'

import React, { type FC, useState, useRef } from 'react'
import ReactSelect from 'react-tailwindcss-select'

interface Option {
  value: string
  label: string
  disabled?: boolean
  isSelected?: boolean
}

const classNames = {
  menuButton: ({ isDisabled }) =>
    `flex text-sm border rounded shadow-sm transition-all duration-300 ${
      isDisabled
        ? 'bg-cool-grey-200 text-cool-grey-500 dark:bg-dark-grey-300 dark:text-dark-grey-900 cursor-not-allowed'
        : 'bg-cool-grey-50 dark:bg-dark-grey-200 [&:user-invalid]:border-red-600 [&:user-invalid]:dark:border-red-600 focus:outline outline-1 outline-primary-500 dark:outline-primary-400'
    }`,
  menu: 'absolute z-10 w-full bg-cool-grey-50 dark:bg-dark-grey-200 shadow-sm border rounded py-1 mt-1.5 text-sm',
  list: 'flex flex-col gap-1 max-h-72 overflow-y-auto',
  listItem: ({ isSelected }) =>
    `transition duration-200 px-2 py-1 -mx-1.5 cursor-pointer select-none truncate rounded text-base ${
      isSelected
        ? 'text-white bg-primary-600'
        : 'hover:bg-black/5 dark:hover:bg-white/5'
    }`,
}

export const Select: FC<{
  defaultValue?: string
  name: string
  options: Option[]
  required?: boolean
}> = ({ defaultValue = '', name, options, required = false }) => {
  const ref = useRef()
  const [value, setValue] = useState<Option>(
    options.find((o) => o.value === defaultValue) || options[0]
  )

  return (
    <>
      <input
        type="hidden"
        name={name}
        value={value.value}
        required={required}
      />
      <ReactSelect
        value={value}
        onChange={(e) => {
          setValue(e as any)
        }}
        options={options as any}
        primaryColor={''}
        classNames={classNames}
      />
    </>
  )
}
