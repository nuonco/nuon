import { getStatusTheme, type TStatusTheme } from '@/utils/status-utils'

export interface GraphAccent {
  border: string
  headerBg: string
  text: string
  dot: string
  pill: string
}

const HUES = {
  blue: {
    border: 'border-blue-300 dark:border-blue-500/40',
    headerBg: 'bg-blue-50 dark:bg-blue-500/10',
    text: 'text-blue-700 dark:text-blue-300',
    dot: 'bg-blue-500',
    pill: 'bg-blue-500/15 text-blue-700 dark:text-blue-300',
  },
  green: {
    border: 'border-green-300 dark:border-green-500/40',
    headerBg: 'bg-green-50 dark:bg-green-500/10',
    text: 'text-green-700 dark:text-green-300',
    dot: 'bg-green-500',
    pill: 'bg-green-500/15 text-green-700 dark:text-green-300',
  },
  amber: {
    border: 'border-amber-300 dark:border-amber-500/40',
    headerBg: 'bg-amber-50 dark:bg-amber-500/10',
    text: 'text-amber-700 dark:text-amber-300',
    dot: 'bg-amber-500',
    pill: 'bg-amber-500/15 text-amber-700 dark:text-amber-300',
  },
  purple: {
    border: 'border-purple-300 dark:border-purple-500/40',
    headerBg: 'bg-purple-50 dark:bg-purple-500/10',
    text: 'text-purple-700 dark:text-purple-300',
    dot: 'bg-purple-500',
    pill: 'bg-purple-500/15 text-purple-700 dark:text-purple-300',
  },
  pink: {
    border: 'border-pink-300 dark:border-pink-500/40',
    headerBg: 'bg-pink-50 dark:bg-pink-500/10',
    text: 'text-pink-700 dark:text-pink-300',
    dot: 'bg-pink-500',
    pill: 'bg-pink-500/15 text-pink-700 dark:text-pink-300',
  },
  cyan: {
    border: 'border-cyan-300 dark:border-cyan-500/40',
    headerBg: 'bg-cyan-50 dark:bg-cyan-500/10',
    text: 'text-cyan-700 dark:text-cyan-300',
    dot: 'bg-cyan-500',
    pill: 'bg-cyan-500/15 text-cyan-700 dark:text-cyan-300',
  },
  red: {
    border: 'border-red-300 dark:border-red-500/40',
    headerBg: 'bg-red-50 dark:bg-red-500/10',
    text: 'text-red-700 dark:text-red-300',
    dot: 'bg-red-500',
    pill: 'bg-red-500/15 text-red-700 dark:text-red-300',
  },
  grey: {
    border: 'border-cool-grey-300 dark:border-dark-grey-600',
    headerBg: 'bg-cool-grey-50 dark:bg-dark-grey-800',
    text: 'text-cool-grey-600 dark:text-cool-grey-300',
    dot: 'bg-cool-grey-400',
    pill: 'bg-cool-grey-500/15 text-cool-grey-600 dark:text-cool-grey-300',
  },
} satisfies Record<string, GraphAccent>

const GROUP_HUES = ['blue', 'green', 'amber', 'purple', 'pink', 'cyan'] as const

export const groupAccent = (index: number): GraphAccent =>
  HUES[GROUP_HUES[index % GROUP_HUES.length]]

const THEME_HUES: Record<TStatusTheme, keyof typeof HUES> = {
  success: 'green',
  error: 'red',
  info: 'blue',
  warn: 'amber',
  neutral: 'grey',
  brand: 'purple',
}

export const statusAccent = (status?: string): GraphAccent =>
  HUES[THEME_HUES[getStatusTheme(status ?? 'pending')]]
