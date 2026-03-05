import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'

export const ResourcesDetails = ({
  resources,
}: {
  resources: Record<string, any>
}) => {
  return (
    <div className="flex flex-col gap-2">
      <Text weight="strong">Resources ({Object.keys(resources)?.length})</Text>

      <Card className="!p-0 !gap-0 divide-y">
        {Object.entries(resources).map(([name, resource]: [string, any]) => (
          <Expand
            id={name}
            key={name}
            headerClassName="px-6"
            heading={<ResourceHeading name={name} kind={resource?.Kind} />}
          >
            <ResourceDetails content={resource?.Content} />
          </Expand>
        ))}
      </Card>
    </div>
  )
}

const ResourceHeading = ({ name, kind }) => {
  return (
    <div className="flex items-center justify-between w-full">
      <Text weight="strong">{name}</Text>
      <Badge
        variant="code"
        size="sm"
        theme={
          kind === 'Deployment'
            ? 'info'
            : kind === 'Service'
              ? 'success'
              : 'neutral'
        }
      >
        {kind}
      </Badge>
    </div>
  )
}

const ResourceDetails = ({ content }: { content: string }) => {
  return (
    <div className="bg-black/2 dark:bg-white/2 border-t relative">
      <div className="absolute top-2 right-2">
        <ClickToCopyButton textToCopy={content} />
      </div>
      <CodeBlock language="yml">{content}</CodeBlock>
    </div>
  )
}
