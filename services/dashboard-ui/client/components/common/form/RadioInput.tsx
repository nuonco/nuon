import { type InputHTMLAttributes, type ReactNode } from 'react'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface IRadioInput
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  labelProps: Omit<ILabel, 'children'> & {
    labelText: string | ReactNode
    labelTextProps?: Omit<IText, 'children'>
  }
}

export const RadioInput = ({
  className,
  labelProps: {
    className: labelClassName,
    labelText,
    labelTextProps = { variant: 'body' },
    ...labelProps
  },
  disabled,
  ...props
}: IRadioInput) => {
  return (
    <Label
      className={cn(
        'flex items-center gap-2 rounded-md p-2 focus-within:outline-1 focus-within:outline-primary-500',
        {
          'hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer': !disabled,
          'cursor-not-allowed': disabled,
        },
        labelClassName
      )}
      {...labelProps}
    >
      <input
        className={cn('accent-primary-600', {
          'cursor-not-allowed': disabled,
        }, className)}
        disabled={disabled}
        {...props}
        type="radio"
      />
      <Text
        className={cn('!leading-none flex-1', labelTextProps?.className)}
        {...labelTextProps}
      >
        {labelText}
      </Text>
    </Label>
  )
}
