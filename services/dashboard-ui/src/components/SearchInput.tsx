'use client'

import classNames from "classnames";
import React from 'react'
import { MagnifyingGlass, XCircle } from '@phosphor-icons/react'
import { Button } from './Button'

export interface ISearchInput
  extends Omit<
    React.HTMLAttributes<HTMLInputElement>,
    "autoComplete" | "onChange" | "type"
  > {
  placeholder?: string;
  onChange: (val: string) => void;
  onClear?: () => void;
  value: string;
}

export const SearchInput = ({
  className,
  placeholder,
  onChange,
  onClear,
  value,
  ...props
}: ISearchInput) => {
  return (
    <label className="relative w-fit flex">
      <MagnifyingGlass
        className="text-cool-grey-500 dark:text-cool-grey-600 absolute top-2.5 left-2"
      />
      <input
        className={classNames(
          "rounded-md pl-8 pr-3.5 py-1.5 h-[36px] font-sans md:min-w-80 border text-sm bg-white dark:bg-dark-grey-100 placeholder:text-cool-grey-500 dark:placeholder:text-cool-grey-700",
          {
            [`${className}`]: Boolean(className)
          }
        )}
        type="search"
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
          onClick={() => (onClear ? onClear() : onChange(""))}
        >
          <XCircle />
        </Button>
      ) : null}
    </label>
  );
};
