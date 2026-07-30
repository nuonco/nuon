import type { TBuilderStep } from '@/components/runbooks/RunbookBuilder/helpers'

export type TMarkdownCell = {
  key: string
  kind: 'markdown'
  content: string
}

export type TStepCell = {
  key: string
  kind: 'step'
  step: TBuilderStep
}

export type TCell = TMarkdownCell | TStepCell
