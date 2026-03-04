import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Link } from '@/components/old/Link'
import { Text } from '@/components/common/Text'

export const StackLinks = ({
  template_url,
  quick_link_url,
}: {
  template_url: string
  quick_link_url: string
}) => {
  return (
    <>
      <div className="border rounded-md shadow p-2 flex flex-col gap-1">
        <span className="flex justify-between items-center">
          <Text variant="body" weight="strong">
            Install quick link
          </Text>
          <ClickToCopyButton textToCopy={quick_link_url} />
        </span>
        <Link href={quick_link_url} target="_blank" rel="noopener noreferrer">
          <Code>{quick_link_url}</Code>
        </Link>
      </div>

      <div className="border rounded-md shadow p-2 flex flex-col gap-1 mt-3">
        <span className="flex justify-between items-center">
          <Text variant="body" weight="strong">
            Install template link
          </Text>
          <ClickToCopyButton textToCopy={template_url} />
        </span>
        <Link href={template_url} target="_blank" rel="noopener noreferrer">
          <Code>{template_url}</Code>
        </Link>
      </div>
    </>
  )
}
