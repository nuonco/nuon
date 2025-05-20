import classNames from 'classnames'
import React, { type FC } from 'react'
import './Text.css'

export type TTextFamily = 'sans' | 'mono'
export type TTextVariant =
  | 'h1'
  | 'h2'
  | 'h3'
  | 'base'
  | 'body'
  | 'subtext'
  | 'label'
export type TTextWeight = 'normal' | 'strong' | 'stronger'

export interface IText extends React.HTMLAttributes<HTMLSpanElement> {
  family?: TTextFamily
  level?: 1 | 2 | 3 | 4 | 5 | 6
  role?: 'paragraph' | 'heading' | 'code' | 'time'
  variant?: TTextVariant
  weight?: TTextWeight
}

export const Text: FC<IText> = ({
  className,
  children,
  family = 'sans',
  level,
  role = 'paragraph',
  variant = 'base',
  weight = 'normal',
  ...props
}) => {
  return (
    <span
      aria-level={role === 'heading' && level ? level : undefined}
      className={classNames(`${variant} ${family} ${weight}`, {
        [`${className}`]: Boolean(className),
      })}
      role={role}
      {...props}
    >
      {children}
    </span>
  )
}
