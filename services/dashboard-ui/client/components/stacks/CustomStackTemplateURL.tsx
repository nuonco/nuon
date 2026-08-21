import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TCustomNestedStack } from '@/types'

// A custom nested stack's template_url is whatever the vendor wrote in their app
// config — usually a path relative to that config's directory, which resolves to
// nothing from the dashboard. template_source_url is the copy Nuon uploaded to
// the install templates bucket, and is the only form that is fetchable here.
// Install-level overrides skip the upload and store an absolute URL directly.
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
        <Link href={href} isExternal>
          {templateURL}
        </Link>
      ) : (
        templateURL
      )}
    </Text>
  )
}
