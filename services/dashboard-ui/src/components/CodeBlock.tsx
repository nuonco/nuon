'use client'

import cn from 'classnames'
import { useEffect, useState } from 'react'
import { Prism } from 'react-syntax-highlighter'
import {
  oneDark,
  oneLight,
} from 'react-syntax-highlighter/dist/cjs/styles/prism'

export function useSystemTheme(): 'dark' | 'light' {
  const [theme, setTheme] = useState<'dark' | 'light'>(() => {
    if (typeof window === 'undefined') return 'light'
    return window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'
  })

  useEffect(() => {
    if (typeof window === 'undefined') return

    const matcher = window.matchMedia('(prefers-color-scheme: dark)')
    const update = () => setTheme(matcher.matches ? 'dark' : 'light')
    matcher.addEventListener('change', update)
    // Initialize on mount
    update()

    return () => matcher.removeEventListener('change', update)
  }, [])

  return theme
}

const DIFF_CLASSES = {
  added:
    'bg-[#F4FBF7] text-green-800 !border-green-400 dark:bg-[#0C1B14] dark:!border-green-500/40 dark:text-green-500 block w-full',
  removed:
    'bg-[#FEF2F2] text-red-800 !border-red-300 dark:bg-[#290C0D] dark:!border-red-500/40 dark:text-red-500 block w-full',
  afterApply: '!italic opacity-70',
}

interface ICodeBlock
  extends Omit<React.HTMLAttributes<HTMLPreElement>, 'children'> {
  children: string
  language:
    | 'json'
    | 'yaml'
    | 'yml'
    | 'hcl'
    | 'sh'
    | 'bash'
    | 'toml'
    | 'markdown'
    | 'md'
    | string
  isDiff?: boolean
  showLineNumbers?: boolean
}

export function CodeBlock({
  className,
  children,
  language,
  isDiff = false,
  showLineNumbers = false,
}: ICodeBlock) {
  const colorScheme = useSystemTheme()
  const theme = colorScheme === 'dark' ? oneDark : oneLight

  return (
    <Prism
      className={cn(
        '!m-0 !p-4 !text-sm !rounded-md !shadow-sm min-h-[3rem] max-h-[40rem] overflow-auto',
        className
      )}
      language={language}
      style={theme}
      wrapLines
      showLineNumbers={showLineNumbers || isDiff}
      lineProps={(lineNumber: number) => {
        if (typeof lineNumber !== 'number') return {}
        const lines = children.split('\n')
        const line = lines[lineNumber - 1] || ''
        let className = ''

        if (isDiff) {
          if (line.startsWith('+')) {
            className = DIFF_CLASSES.added
          } else if (line.startsWith('-')) {
            className = DIFF_CLASSES.removed
          }
        }

        if (line.includes('"Known after apply"')) {
          className = className
            ? `${className} ${DIFF_CLASSES.afterApply}`
            : DIFF_CLASSES.afterApply
        }

        return className ? { className } : {}
      }}
      codeTagProps={{
        className: ' font-mono w-full',
      }}
      customStyle={{
        fontFamily: 'var(--font-hack)',
      }}
    >
      {children}
    </Prism>
  )
}
