import type { InputHTMLAttributes } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface ICheckbox
  extends InputHTMLAttributes<Omit<HTMLInputElement, 'type'>> {}

export const Checkbox = ({ className, ...props }: ICheckbox) => {
  return (
    <input
      type="checkbox"
      className={cn(
        'focus:ring-primary-500 focus:border-primary-500 accent-primary-600',
        className
      )}
      {...props}
    />
  )
}

export interface ICheckboxInput extends ICheckbox {
  labelProps: Omit<ILabel, 'children'> & {
    labelText: string
    labelTextProps?: Omit<IText, 'children'>
  }
}

export const CheckboxInput = ({
  labelProps: {
    className: labelClassName,
    labelText,
    labelTextProps = { variant: 'body' },
    ...labelProps
  },
  ...props
}: ICheckboxInput) => {
  return (
    <Label
      className={cn(
        'flex items-center gap-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-md p-2 focus-within:outline-1 focus-within:outline-primary-500 cursor-pointer ',
        labelClassName
      )}
      {...labelProps}
    >
      <Checkbox {...props} />
      <Text
        className={cn('!leading-none', labelTextProps?.className)}
        {...labelTextProps}
      >
        {labelText}
      </Text>
    </Label>
  )
}

export interface ICheckboxInputWithButton extends ICheckbox {
  buttonProps: IButtonAsButton
  checkboxClassName?: string
}

export const CheckboxInputWithButton = ({
  buttonProps: { className: buttonClassName, children, ...buttonProps },
  className,
  checkboxClassName,
  ...props
}: ICheckboxInputWithButton) => {
  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Checkbox className={checkboxClassName} {...props} />
      <Button
        className={cn(
          '!p-1 flex items-center justify-between group w-full',
          buttonClassName
        )}
        variant="ghost"
        size="sm"
        {...buttonProps}
      ></Button>
    </div>
  )
}
