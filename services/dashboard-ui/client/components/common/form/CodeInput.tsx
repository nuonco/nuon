import { type TextareaHTMLAttributes, useState, useMemo } from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { json } from '@codemirror/lang-json'
import { StreamLanguage } from '@codemirror/language'
import { shell } from '@codemirror/legacy-modes/mode/shell'
import { EditorView } from '@codemirror/view'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { useSystemTheme } from '@/hooks/use-system-theme'
import { cn } from '@/utils/classnames'

export interface ICodeInput
  extends Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'size' | 'onChange'> {
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
  language?: 'json' | 'shell' | 'bash'
  minHeight?: number
  onChange?: (e: { target: { value: string; name?: string } }) => void
}

function getLanguageExtension(language: string) {
  switch (language) {
    case 'json':
      return json()
    case 'shell':
    case 'bash':
      return StreamLanguage.define(shell)
    default:
      return json()
  }
}

const bgOverride = EditorView.theme({
  '&': {
    backgroundColor: 'var(--bg-code)',
  },
  '.cm-gutters': {
    backgroundColor: 'var(--bg-code)',
  },
})

export const CodeInput = ({
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
  name,
  ...props
}: ICodeInput) => {
  const colorScheme = useSystemTheme()
  const [internalValue, setInternalValue] = useState(
    (defaultValue as string) || ''
  )
  const currentValue = value !== undefined ? (value as string) : internalValue

  const handleChange = (val: string) => {
    if (value === undefined) {
      setInternalValue(val)
    }
    onChange?.({ target: { value: val, name } })
  }

  const extensions = useMemo(() => {
    const lang = getLanguageExtension(language)
    return lang ? [lang, bgOverride] : [bgOverride]
  }, [language])

  const fontSize = size === 'sm' ? '12px' : size === 'lg' ? '16px' : '14px'

  const baseClasses = cn(
    'rounded-md border transition-colors duration-200 overflow-hidden',
    'shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)]',
    'focus-within:!border-primary-500 dark:focus-within:!border-primary-500',
    {
      'border-cool-grey-500/24 dark:border-cool-grey-500/24':
        !error && !disabled,
      '!border-red-500 dark:!border-red-400': error,
      '!border-cool-grey-300 dark:!border-dark-grey-600': disabled,
      '!shadow-none': disabled,
      'opacity-50 cursor-not-allowed': disabled,
      'focus-within:!border-cool-grey-300 dark:focus-within:!border-dark-grey-600':
        disabled,
    },
    className
  )

  const codeEditor = (
    <div className={baseClasses}>
      {name && <input type="hidden" name={name} value={currentValue} />}
      <CodeMirror
        value={currentValue}
        onChange={handleChange}
        extensions={extensions}
        readOnly={!!disabled}
        editable={!disabled}
        theme={colorScheme}
        minHeight={`${minHeight}px`}
        basicSetup={{
          lineNumbers: true,
          foldGutter: false,
          highlightActiveLine: true,
          bracketMatching: true,
          indentOnInput: true,
          tabSize: 2,
        }}
        style={{ fontSize, fontFamily: 'var(--font-hack)' }}
        aria-invalid={error}
        aria-describedby={
          helperText || errorMessage ? `${props.id}-description` : undefined
        }
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
