import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface IDestructiveChange {
  id: string
  title: string
  operation: 'delete' | 'replace'
}

export const destructiveChanges = (
  sections: { id: string; title: string; operation: string }[]
): IDestructiveChange[] =>
  sections.flatMap((section) =>
    section.operation === 'delete' || section.operation === 'replace'
      ? [
          {
            id: section.id,
            title: section.title,
            operation: section.operation,
          },
        ]
      : []
  )

const OPERATION_LABELS = {
  delete: 'Delete',
  replace: 'Replace',
} as const

const OPERATION_CLASSES = {
  delete: 'text-diff-remove',
  replace: 'text-fuchsia-600 dark:text-fuchsia-300',
} as const

const headingFor = (changes: IDestructiveChange[]) => {
  const deletes = changes.some((change) => change.operation === 'delete')
  const replaces = changes.some((change) => change.operation === 'replace')
  const verb =
    deletes && replaces ? 'delete or replace' : deletes ? 'delete' : 'replace'
  const noun = changes.length === 1 ? 'a resource' : 'resources'
  return `Approving this plan will ${verb} ${noun}.`
}

export const DestructiveChangeBanner = ({
  changes,
}: {
  changes: IDestructiveChange[]
}) => {
  if (!changes.length) return null
  const replaces = changes.some((change) => change.operation === 'replace')

  return (
    <Banner theme="warn">
      <div className="flex flex-col gap-2">
        <Text weight="strong">{headingFor(changes)}</Text>
        {replaces ? (
          <Text variant="subtext">
            Replace destroys the resource and creates it again.
          </Text>
        ) : null}
        <Code className="!min-h-0 gap-1 !p-3">
          {changes.map((change) => (
            <span
              key={change.id}
              className="flex min-w-0 items-baseline gap-2"
            >
              <Text
                weight="stronger"
                className={cn('shrink-0', OPERATION_CLASSES[change.operation])}
              >
                {OPERATION_LABELS[change.operation]}
              </Text>
              <Text family="mono" className="min-w-0 break-all">
                {change.title}
              </Text>
            </span>
          ))}
        </Code>
      </div>
    </Banner>
  )
}
