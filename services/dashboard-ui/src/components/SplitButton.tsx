'use client'

import classNames from 'classnames'
import { type ReactNode, type MouseEvent } from 'react'
import { DotsThreeVerticalIcon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown, type IDropdown } from '@/components/Dropdown'

interface ISplitButton extends Omit<IDropdown, 'text' | 'noIcon' | 'variant'> {
  buttonText: ReactNode
  buttonOnClick?: (event: MouseEvent<HTMLButtonElement>) => void
  variant?: 'default' | 'primary' | 'danger' | 'caution'
}

export const SplitButton = ({
  className,
  children,
  buttonText,
  buttonOnClick,
  variant = 'default',
  ...props
}: ISplitButton) => {
  return (
    <div
      className={classNames(
        'border rounded-md w-fit flex items-center divide-x',
        {
          [`${className}`]: Boolean(className),
        }
      )}
    >
      <Button
        className="!rounded-e-none !border-none"
        variant={variant}
        onClick={buttonOnClick}
      >
        {buttonText}
      </Button>
      <Dropdown
        variant={variant === 'caution' ? 'default' : variant}
        className="!p-2 !rounded-s-none !border-none"
        text={<DotsThreeVerticalIcon size="14" />}
        noIcon
        {...props}
      >
        {children}
      </Dropdown>
    </div>
  )
}
