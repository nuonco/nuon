import { Text } from '@/components/common/Text'
import { LabeledValue } from '@/components/common/LabeledValue'

interface IComparisonStep {
  metadata: Record<string, any>
  status?: string
}

export const ComparisonStep = ({ metadata, status }: IComparisonStep) => {
  const skipReason = metadata.skip_reason as string | undefined
  const baseSHA = metadata.base_sha as string | undefined
  const headSHA = metadata.head_sha as string | undefined
  const filesChanged = metadata.files_changed as number | undefined
  const additions = metadata.additions as number | undefined
  const removals = metadata.removals as number | undefined

  if (skipReason) {
    return (
      <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
        <Text variant="base">Comparison skipped: {skipReason}</Text>
      </div>
    )
  }

  if (status === 'pending' || status === 'in-progress') {
    return (
      <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
        <Text variant="base" theme="neutral">
          Computing run comparison…
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {(baseSHA || headSHA) && (
        <div className="grid grid-cols-2 gap-3">
          {baseSHA && (
            <LabeledValue label="Base commit">
              <Text variant="subtext" family="mono">
                {baseSHA}
              </Text>
            </LabeledValue>
          )}
          {headSHA && (
            <LabeledValue label="Head commit">
              <Text variant="subtext" family="mono">
                {headSHA}
              </Text>
            </LabeledValue>
          )}
        </div>
      )}
      {(filesChanged != null || additions != null || removals != null) && (
        <div className="grid grid-cols-3 gap-3">
          {filesChanged != null && (
            <LabeledValue label="Files changed">{String(filesChanged)}</LabeledValue>
          )}
          {additions != null && (
            <LabeledValue label="Additions">{String(additions)}</LabeledValue>
          )}
          {removals != null && (
            <LabeledValue label="Removals">{String(removals)}</LabeledValue>
          )}
        </div>
      )}
    </div>
  )
}
