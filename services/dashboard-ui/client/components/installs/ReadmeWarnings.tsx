import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { toSentenceCase } from '@/utils/string-utils'

export const ReadmeWarnings = ({ warnings }: { warnings?: string[] }) => {
  if (!warnings?.length) return null

  return (
    <Banner theme="warn">
      <div className="flex flex-col gap-2">
        <Text weight="strong">Template rendering incomplete</Text>
        <Text variant="subtext">
          Some dynamic values in this README could not be resolved. This
          typically means the referenced data isn't available yet — for example,
          outputs from a component that hasn't finished deploying. These values
          will populate automatically once the data is ready.
        </Text>
        <ul className="flex flex-col gap-1 mt-1 text-sm list-disc list-inside">
          {warnings.map((warn, i) => (
            <li key={i}>{toSentenceCase(warn)}</li>
          ))}
        </ul>
      </div>
    </Banner>
  )
}
