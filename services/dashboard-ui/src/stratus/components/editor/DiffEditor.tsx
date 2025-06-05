'use client'

import classNames from 'classnames'
import dynamic from 'next/dynamic'
import React, { type FC } from 'react'

const MonacoDiffEditor = dynamic(
  () => import('@monaco-editor/react').then((m) => m.DiffEditor),
  { ssr: false }
)

type MonacoDiffEditorProps = React.ComponentProps<typeof MonacoDiffEditor>

interface IDiffEditor extends MonacoDiffEditorProps {
  wrapperClassName?: string
}

export const DiffEditor: FC<IDiffEditor> = ({
  height = 500,
  options = {
    renderSideBySide: true,
    readOnly: true,
    fontFamily: '__hack_f5efd2',
    fontSize: 12,
  },
  theme = 'vs-dark',
  wrapperClassName,
  ...props
}) => {
  return (
    <div
      className={classNames('rounded-md min-h-[500px] overflow-hidden', {
        [`${wrapperClassName}`]: Boolean(wrapperClassName),
      })}
    >
      <MonacoDiffEditor
        height={height}
        theme={theme}
        options={options}
        {...props}
      />
    </div>
  )
}
