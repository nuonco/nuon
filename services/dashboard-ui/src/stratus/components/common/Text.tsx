import React from 'react'
import { cn } from '@/stratus/components/helpers'
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
export type TTextTheme = 'default' | 'muted' | 'highlighted'

export interface IText extends React.HTMLAttributes<HTMLSpanElement> {
  family?: TTextFamily
  level?: 1 | 2 | 3 | 4 | 5 | 6
  role?: 'paragraph' | 'heading' | 'code' | 'time'
  theme?: TTextTheme
  variant?: TTextVariant
  weight?: TTextWeight
}

export const Text = ({
  className,
  children,
  family = 'sans',
  level,
  role = 'paragraph',
  variant = 'body',
  theme = 'default',
  weight = 'normal',
  ...props
}: IText) => {
  return (
    <span
      aria-level={role === 'heading' && level ? level : undefined}
      className={cn(variant, family, weight, theme, className)}
      role={role}
      {...props}
    >
      {children}
    </span>
  )
}
