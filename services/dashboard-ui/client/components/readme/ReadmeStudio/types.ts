export type TConditionOp = 'exists' | 'not-exists' | 'eq' | 'ne'

export type TCondition = {
  path: string
  op: TConditionOp
  value?: string
}

export type TValueKind = 'text' | 'status' | 'time'

export type TStatusRowItem = {
  key: string
  label: string
  kind: TValueKind
  path: string
}

export type TTableColumn = {
  key: string
  header: string
  kind: TValueKind
  path: string
}

export type TMarkdownBlock = {
  key: string
  type: 'markdown'
  content: string
}

export type TBannerBlock = {
  key: string
  type: 'banner'
  theme: 'info' | 'warn' | 'error' | 'success'
  content: string
  condition?: TCondition
}

export type TStatusRowBlock = {
  key: string
  type: 'status-row'
  items: TStatusRowItem[]
}

export type TTableBlock = {
  key: string
  type: 'table'
  sourcePath: string
  limit?: number
  emptyText: string
  columns: TTableColumn[]
}

export type TRawBlock = {
  key: string
  type: 'raw'
  content: string
}

export type TRunbookBlock = {
  key: string
  type: 'runbook'
  id: string
  name: string
}

export type TActionBlock = {
  key: string
  type: 'action'
  id: string
  name: string
}

export type TComponentBlock = {
  key: string
  type: 'component'
  id: string
  name: string
}

export type TBlock =
  | TMarkdownBlock
  | TBannerBlock
  | TStatusRowBlock
  | TTableBlock
  | TRunbookBlock
  | TActionBlock
  | TComponentBlock
  | TRawBlock

export type TBlockType = TBlock['type']

export type TEntityOption = {
  id: string
  name: string
}

export type TStateVariable = {
  template: string
  value?: string
}

export type TArraySource = {
  path: string
  keys: string[]
  length: number
}
