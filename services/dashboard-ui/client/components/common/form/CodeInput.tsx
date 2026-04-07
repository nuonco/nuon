import { type TextareaHTMLAttributes, forwardRef, useState } from 'react'
import CodeEditor from '@uiw/react-textarea-code-editor'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface ICodeInput
  extends Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'size'> {
  labelProps?: Omit<ILabel, 'children'> & {
    labelText: string
    labelTextProps?: Omit<IText, 'children'>
  }
  helperText?: string
  helperTextProps?: Omit<IText, 'children'>
  error?: boolean
  errorMessage?: string
  errorMessageProps?: Omit<IText, 'children'>
  size?: 'sm' | 'md' | 'lg'
  language?: 'json' | 'yaml' | 'javascript' | 'typescript' | 'shell' | 'toml' | 'hcl'
  minHeight?: number
}

export const CodeInput = forwardRef<HTMLTextAreaElement, ICodeInput>(
  (
    {
      className,
      labelProps,
      helperText,
      helperTextProps = { variant: 'subtext' },
      error,
      errorMessage,
      errorMessageProps = { variant: 'subtext', theme: 'error' },
      size = 'md',
      language = 'json',
      minHeight = 120,
      disabled,
      value,
      defaultValue,
      onChange,
      ...props
    },
    ref
  ) => {
    const [internalValue, setInternalValue] = useState(defaultValue || '')
    const currentValue = value !== undefined ? value : internalValue

    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      if (value === undefined) {
        setInternalValue(e.target.value)
      }
      onChange?.(e)
    }

    const paddingClasses = {
      sm: 12,
      md: 16,
      lg: 20,
    }

    const baseClasses = cn(
      'rounded-md border transition-colors duration-200 overflow-auto',
      'shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)]',
      'focus-within:!border-primary-500 dark:focus-within:!border-primary-500',
      'user-invalid:!border-red-500 user-invalid:dark:!border-red-400',
      'user-invalid:focus-within:!border-red-500',
      {
        'border-cool-grey-500/24 dark:border-cool-grey-500/24': !error && !disabled,

        '!border-red-500 dark:!border-red-400': error,

        '!border-cool-grey-300 dark:!border-dark-grey-600': disabled,
        '!shadow-none': disabled,
        'opacity-50 cursor-not-allowed': disabled,
        'focus-within:!border-cool-grey-300 dark:focus-within:!border-dark-grey-600': disabled,
      },
      className
    )

    const codeEditor = (
      <div className={baseClasses}>
        <CodeEditor
          ref={ref}
          value={currentValue as string}
          language={language}
          onChange={handleChange}
          padding={paddingClasses[size]}
          readOnly={disabled}
          style={{
            backgroundColor: 'light-dark(#ffffff, #141217)',
            color: 'light-dark(#141217, #ffffff)',
            fontFamily:
              'ui-monospace, SFMono-Regular, "SF Mono", Monaco, Menlo, Consolas, "Liberation Mono", "Courier New", monospace',
            fontSize: size === 'sm' ? '12px' : size === 'lg' ? '16px' : '14px',
            lineHeight: '1.5',
            minWidth: '100%',
            width: 'max-content',
            minHeight: `${minHeight}px`,
            resize: 'vertical',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            overflowWrap: 'break-word',
          }}
          autoCapitalize="none"
          spellCheck={false}
          aria-invalid={error}
          aria-describedby={
            helperText || errorMessage ? `${props.id}-description` : undefined
          }
          {...props}
        />
      </div>
    )

    const renderDescription = () => {
      if (error && errorMessage) {
        return (
          <Text
            id={`${props.id}-description`}
            className={cn('block', errorMessageProps?.className)}
            {...errorMessageProps}
          >
            {errorMessage}
          </Text>
        )
      }

      if (helperText) {
        return (
          <Text
            id={`${props.id}-description`}
            className={cn('block', helperTextProps?.className)}
            {...helperTextProps}
          >
            {helperText}
          </Text>
        )
      }

      return null
    }

    if (labelProps) {
      const { labelText, labelTextProps, ...restLabelProps } = labelProps
      return (
        <div className="flex flex-col gap-1">
          <Label
            className={cn('block', labelProps.className)}
            htmlFor={props.id}
            {...restLabelProps}
          >
            <Text
              className={cn('font-medium', labelTextProps?.className)}
              variant="body"
              {...labelTextProps}
            >
              {labelText}
            </Text>
          </Label>
          {codeEditor}
          {renderDescription()}
        </div>
      )
    }

    return (
      <div className="flex flex-col gap-1">
        {codeEditor}
        {renderDescription()}
      </div>
    )
  }
)

CodeInput.displayName = 'CodeInput'
