import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TCustomNestedStack } from '@/types'

export const CustomStackTemplateURL = ({
  stack,
}: {
  stack?: TCustomNestedStack
}) => {
  const templateURL = stack?.template_url
  if (!templateURL) return null

  const href =
    stack?.template_source_url ||
    (/^https?:\/\//.test(templateURL) ? templateURL : undefined)

  return (
    <Text variant="subtext">
      {href ? (
        <Link href={href} isExternal variant="inline">
          {templateURL}
        </Link>
      ) : (
        templateURL
      )}
    </Text>
  )
}
