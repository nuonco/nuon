import { useMemo } from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { json } from '@codemirror/lang-json'
import { StreamLanguage } from '@codemirror/language'
import { shell } from '@codemirror/legacy-modes/mode/shell'
import { EditorView } from '@codemirror/view'
import { useSystemTheme } from '@/hooks/use-system-theme'
import { cn } from '@/utils/classnames'

export type TEditorLanguage = 'bash' | 'json' | 'sh' | 'shell'

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
}

function getLanguageExtension(language: TEditorLanguage) {
  switch (language) {
    case 'json':
      return json()
    case 'bash':
    case 'sh':
    case 'shell':
      return StreamLanguage.define(shell)
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

export const Editor = ({
  value = '',
  onChange,
  language = 'bash',
  placeholder = 'Enter code here...',
  disabled = false,
  readOnly = false,
  name,
  className,
  minHeight = 200,
  maxHeight = 600,
}: IEditor) => {
  const colorScheme = useSystemTheme()

  const extensions = useMemo(() => {
    const lang = getLanguageExtension(language)
    return lang ? [lang, bgOverride] : [bgOverride]
  }, [language])

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
      {name && <input type="hidden" name={name} value={value} />}
      <CodeMirror
        value={value}
        onChange={onChange}
        extensions={extensions}
        placeholder={placeholder}
        readOnly={readOnly || disabled}
        editable={!disabled}
        theme={colorScheme}
        minHeight={`${minHeight}px`}
        maxHeight={`${maxHeight}px`}
        basicSetup={{
          lineNumbers: true,
          foldGutter: false,
          highlightActiveLine: true,
          bracketMatching: true,
          indentOnInput: true,
          tabSize: 2,
        }}
        style={{
          fontSize: '14px',
          fontFamily: 'var(--font-hack)',
        }}
      />
    </div>
  )
}
