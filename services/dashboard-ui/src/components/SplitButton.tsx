'use client'

import classNames from 'classnames'
import { type ReactNode, type MouseEvent } from 'react'
import { DotsThreeVerticalIcon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown, type IDropdown } from '@/components/Dropdown'

interface ISplitButton extends Omit<IDropdown, 'text' | 'noIcon' | 'variant'> {
  buttonText: ReactNode
  buttonIcon?: ReactNode
  buttonOnClick?: (event: MouseEvent<HTMLButtonElement>) => void
  buttonClassName?: string
  dropdownClassName?: string
  variant?: 'default' | 'primary' | 'danger' | 'caution'
}

export const SplitButton = ({
  className,
  children,
  buttonText,
  buttonIcon,
  buttonOnClick,
  buttonClassName,
  dropdownClassName,
  disabled = false,
  variant = 'default',
  ...props
}: ISplitButton) => {
  return (
    <div
      className={classNames('border rounded-md w-fit flex items-center', {
        [`${className}`]: Boolean(className),
      })}
    >
      <Button
        disabled={disabled}
        className={classNames(
          '!rounded-e-none !border-0 !flex !h-[32px] !bg-inherit',
          {
            [`${buttonClassName}`]: Boolean(buttonClassName),
          }
        )}
        variant={variant}
        onClick={buttonOnClick}
      >
        <span className="flex items-center gap-2">
          {buttonIcon}
          {buttonText}
        </span>
      </Button>
      <Dropdown
        disabled={disabled}
        variant={variant === 'caution' ? 'default' : variant}
        className={classNames(
          '!p-2 !rounded-s-none !border-r-0 !border-y-0 !flex !h-[32px] !bg-inherit',
          {
            [`${dropdownClassName}`]: Boolean(dropdownClassName),
          }
        )}
        text={<DotsThreeVerticalIcon size="14" />}
        noIcon
        {...props}
      >
        {children}
      </Dropdown>
    </div>
  )
}
