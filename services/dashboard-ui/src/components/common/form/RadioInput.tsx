import { type InputHTMLAttributes } from 'react'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface IRadioInput
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  labelProps: Omit<ILabel, 'children'> & {
    labelText: string
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
  ...props
}: IRadioInput) => {
  return (
    <Label
      className={cn(
        'flex items-center gap-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-md p-2 focus-within:outline-1 focus-within:outline-primary-500 cursor-pointer ',
        labelClassName
      )}
      {...labelProps}
    >
      <input
        className={cn('accent-primary-600', className)}
        {...props}
        type="radio"
      />
      <Text
        className={cn('!leading-none', labelTextProps?.className)}
        {...labelTextProps}
      >
        {labelText}
      </Text>
    </Label>
  )
}
