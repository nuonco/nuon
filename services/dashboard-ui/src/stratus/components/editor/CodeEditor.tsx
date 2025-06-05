'use client'

import classNames from 'classnames'
import dynamic from 'next/dynamic'
import React, { FC } from 'react'

const MonacoEditor = dynamic(() => import('@monaco-editor/react'), {
  ssr: false,
})

type MonacoEditorProps = React.ComponentProps<typeof MonacoEditor>

interface ICodeEditor extends MonacoEditorProps {
  wrapperClassName?: string
}

export const CodeEditor: FC<ICodeEditor> = ({
  height = 500,
  options = {
    codeLens: false,
    fontSize: 12,
    fontFamily: '__hack_f5efd2',
    readOnly: true,
    minimap: { enabled: false },
  },
  theme = 'light',
  wrapperClassName,
  ...props
}) => (
  <div
    className={classNames('rounded-md min-h-[500px] overflow-hidden', {
      [`${wrapperClassName}`]: Boolean(wrapperClassName),
    })}
  >
    <MonacoEditor height={height} theme={theme} options={options} {...props} />
  </div>
)
