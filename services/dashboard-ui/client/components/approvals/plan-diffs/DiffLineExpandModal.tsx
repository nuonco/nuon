import { useMemo } from 'react'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import { diffLines } from '@/utils/code-utils'
import { isStringJson } from '@/utils/terraform-utils'
import { TreeDiffValue } from './terraform/TreeDiffValue'

type DiffPrefix = '~' | '+' | '-'

const PREFIX_STYLES: Record<DiffPrefix, string> = {
  '~': 'text-orange-600 dark:text-orange-400',
  '+': 'text-green-600 dark:text-green-400',
  '-': 'text-red-600 dark:text-red-400',
}

function isPlainString(val: unknown): val is string {
  return typeof val === 'string' && !isStringJson(val)
}

function detectLanguage(value: string): string {
  const trimmed = value.trim()
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) return 'json'
  if (trimmed.includes(': ') || trimmed.includes(':\n')) return 'yaml'
  return 'yaml'
}

const DiffLineModal = ({
  label,
  prefix,
  before,
  after,
  ...props
}: IModal & {
  label: string
  prefix: DiffPrefix
  before: unknown
  after: unknown
}) => {
  const isMultiline =
    (typeof before === 'string' && before.includes('\n')) ||
    (typeof after === 'string' && after.includes('\n'))
  const useCodeBlock =
    isMultiline &&
    (isPlainString(before) || before === null || before === undefined) &&
    (isPlainString(after) || after === null || after === undefined)

  const { diff, language } = useMemo(() => {
    if (!useCodeBlock) return { diff: '', language: 'yaml' }
    const b = typeof before === 'string' ? before : ''
    const a = typeof after === 'string' ? after : ''
    return { diff: diffLines(b, a), language: detectLanguage(a || b) }
  }, [before, after, useCodeBlock])

  const heading = (
    <Text variant="h3" weight="strong" family="mono" className={PREFIX_STYLES[prefix]}>
      {prefix} {label}:
    </Text>
  )

  return (
    <Modal heading={heading} size="xl" showFooter={false} {...props}>
      {useCodeBlock ? (
        <div className="bg-code rounded-md overflow-auto max-h-[70vh] [&_pre]:!overflow-visible">
          <CodeBlock
            className="!rounded-none !shadow-none !m-0"
            language={language}
            isDiff
          >
            {diff}
          </CodeBlock>
        </div>
      ) : (
        <div className="bg-code rounded-md overflow-auto max-h-[70vh] p-4">
          <TreeDiffValue before={before} after={after} />
        </div>
      )}
    </Modal>
  )
}

export const DiffLineExpandButton = ({
  label,
  prefix,
  before,
  after,
}: {
  label: string
  prefix: DiffPrefix
  before: unknown
  after: unknown
}) => {
  const { addModal } = useSurfaces()

  return (
    <button
      type="button"
      className="inline-flex items-center ml-2 px-1 py-0.5 rounded text-blue-600 dark:text-blue-400 hover:bg-blue-100 dark:hover:bg-blue-900/30 opacity-60 hover:opacity-100 transition-opacity align-middle"
      onClick={(e) => {
        e.stopPropagation()
        const modal = (
          <DiffLineModal
            label={label}
            prefix={prefix}
            before={before}
            after={after}
          />
        )
        addModal(modal)
      }}
      title="Expand value"
      aria-label={`Expand ${label}`}
    >
      <Icon variant="ArrowsOutIcon" size={12} />
    </button>
  )
}
