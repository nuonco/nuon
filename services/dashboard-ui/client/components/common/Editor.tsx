import { forwardRef } from 'react'
import CodeEditor from '@uiw/react-textarea-code-editor'
import { useSystemTheme } from '@/hooks/use-system-theme'
import { cn } from '@/utils/classnames'

export type TEditorLanguage =
  | 'bash'
  | 'css'
  | 'go'
  | 'hcl'
  | 'html'
  | 'javascript'
  | 'json'
  | 'markdown'
  | 'python'
  | 'sh'
  | 'typescript'
  | 'yaml'

export interface IEditor {
  value?: string
  onChange?: (value: string) => void
  language?: TEditorLanguage
  placeholder?: string
  disabled?: boolean
  readOnly?: boolean
  name?: string
  className?: string
  minHeight?: number
  maxHeight?: number
  padding?: number
}

export const Editor = forwardRef<HTMLTextAreaElement, IEditor>(
  (
    {
      value = '',
      onChange,
      language = 'javascript',
      placeholder = 'Enter code here...',
      disabled = false,
      readOnly = false,
      name,
      className,
      minHeight = 200,
      maxHeight = 600,
      padding = 16,
    },
    ref
  ) => {
    const colorScheme = useSystemTheme()

    const handleChange = (
      event: React.ChangeEvent<HTMLTextAreaElement>
    ): void => {
      if (onChange) {
        onChange(event.target.value)
      }
    }

    return (
      <div
        className={cn(
          'relative rounded-md shadow-sm overflow-hidden',
          {
            'opacity-50 cursor-not-allowed': disabled,
          },
          className
        )}
      >
        <CodeEditor
          ref={ref}
          value={value}
          language={language}
          placeholder={placeholder}
          onChange={handleChange}
          disabled={disabled}
          readOnly={readOnly}
          name={name}
          padding={padding}
          style={{
            minHeight: `${minHeight}px`,
            maxHeight: `${maxHeight}px`,
            overflow: 'auto',
            fontSize: '14px',
            fontFamily: 'var(--font-hack)',
            backgroundColor: 'var(--bg-code)',
          }}
          data-color-mode={colorScheme}
        />
      </div>
    )
  }
)

Editor.displayName = 'Editor'
