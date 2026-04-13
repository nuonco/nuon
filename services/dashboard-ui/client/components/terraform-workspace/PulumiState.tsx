import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Code } from '@/components/common/Code'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { JSONViewer } from '@/components/common/JSONViewer'
import { LabeledValue } from '@/components/common/LabeledValue'
import { SearchInput } from '@/components/common/SearchInput'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'

export type TPulumiResource = {
  urn: string
  type: string
  inputs?: Record<string, any>
  outputs?: Record<string, any>
  parent?: string
  provider?: string
  dependencies?: string[]
}

export type TPulumiState = {
  resources?: TPulumiResource[]
}

function getStackOutputs(state: TPulumiState): Record<string, any> {
  const stackResource = state.resources?.find(
    (r) => r.type === 'pulumi:pulumi:Stack'
  )
  return stackResource?.outputs ?? {}
}

function getUserResources(state: TPulumiState): TPulumiResource[] {
  return (
    state.resources?.filter(
      (r) =>
        r.type !== 'pulumi:pulumi:Stack' &&
        !r.type.startsWith('pulumi:providers:')
    ) ?? []
  )
}

function getResourceName(urn: string): string {
  const parts = urn.split('::')
  return parts[parts.length - 1] ?? urn
}

export const PulumiState = ({ pulumiState }: { pulumiState: TPulumiState }) => {
  const [outputSearch, setOutputSearch] = useState('')
  const [resourceSearch, setResourceSearch] = useState('')

  const outputs = getStackOutputs(pulumiState)
  const resources = getUserResources(pulumiState)

  const filteredOutputEntries = Object.entries(outputs).filter(([key]) =>
    key.toLowerCase().includes(outputSearch.toLowerCase())
  )

  const filteredResources = resources.filter(
    (r) =>
      r.urn?.toLowerCase().includes(resourceSearch.toLowerCase()) ||
      r.type?.toLowerCase().includes(resourceSearch.toLowerCase()) ||
      getResourceName(r.urn)?.toLowerCase().includes(resourceSearch.toLowerCase())
  )

  return (
    <Tabs
      tabs={{
        outputs: (
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-3 mt-4">
              <SearchInput
                placeholder="Search outputs"
                value={outputSearch}
                onChange={setOutputSearch}
              />
              <Text variant="subtext" theme="neutral">
                {filteredOutputEntries.length} of {Object.entries(outputs).length}
              </Text>
            </div>
            <div className="flex flex-col gap-4">
              {filteredOutputEntries.length ? (
                filteredOutputEntries.map(([key, val]) => (
                  <LabeledValue
                    className="border rounded-md p-4"
                    key={key}
                    label={key}
                  >
                    {Array.isArray(val) ? (
                      <CodeBlock language="json">
                        {JSON.stringify(val, null, 2)}
                      </CodeBlock>
                    ) : val !== null && val !== undefined && typeof val === 'object' ? (
                      <JSONViewer
                        data={val}
                        expanded={1}
                        showDataTypes={false}
                        showSize={false}
                        className="!border-0 !rounded-none"
                      />
                    ) : val !== null && val !== undefined ? (
                      <Code className="w-full overflow-x-auto !text-xs !p-2" variant="inline">
                        {String(val)}
                      </Code>
                    ) : (
                      <Text variant="subtext" theme="neutral">
                        —
                      </Text>
                    )}
                  </LabeledValue>
                ))
              ) : (
                <Text className="p-4" variant="body" theme="neutral">
                  No outputs match your search.
                </Text>
              )}
            </div>
          </div>
        ),
        resources: (
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-3 mt-4">
              <SearchInput
                placeholder="Search resources"
                value={resourceSearch}
                onChange={setResourceSearch}
              />
              <Text variant="subtext" theme="neutral">
                {filteredResources.length} of {resources.length}
              </Text>
            </div>
            <div className="flex flex-col gap-4">
              {resources.length ? (
                filteredResources.length ? (
                  filteredResources.map((resource) => (
                    <Expand
                      id={resource.urn}
                      key={resource.urn}
                      className="border rounded-md"
                      headerClassName="!px-4 py-2"
                      heading={<PulumiResourceHeading resource={resource} />}
                    >
                      <PulumiResourceDetails resource={resource} />
                    </Expand>
                  ))
                ) : (
                  <Text className="p-4" variant="body" theme="neutral">
                    No resources match your search.
                  </Text>
                )
              ) : (
                <Text>No Pulumi resources</Text>
              )}
            </div>
          </div>
        ),
      }}
    />
  )
}

const PulumiResourceHeading = ({ resource }: { resource: TPulumiResource }) => {
  const name = getResourceName(resource.urn)
  return (
    <div className="flex flex-col gap-2 w-full">
      <div className="flex items-center justify-between">
        <span className="flex flex-col gap-2 text-left">
          <Text weight="strong">{name}</Text>
          <Text flex className="gap-6" theme="neutral">
            <Text variant="label">
              <b>URN:</b> {resource.urn}
            </Text>
          </Text>
        </span>
        <Badge className="self-start" variant="code" size="sm">
          {resource.type}
        </Badge>
      </div>
    </div>
  )
}

const PulumiResourceDetails = ({ resource }: { resource: TPulumiResource }) => {
  return (
    <div className="p-4 border-t flex flex-col gap-6 max-h-180 overflow-auto">
      {resource.dependencies?.length ? (
        <div className="flex flex-col gap-2">
          <Text>Dependencies</Text>
          <div className="flex items-center gap-2 flex-wrap">
            {resource.dependencies.map((d) => (
              <Badge variant="code" size="sm" key={d}>
                {getResourceName(d)}
              </Badge>
            ))}
          </div>
        </div>
      ) : null}

      {resource.inputs && Object.keys(resource.inputs).length > 0 && (
        <div className="flex flex-col gap-2">
          <Text>Inputs</Text>
          <CodeBlock language="json">
            {JSON.stringify(resource.inputs, null, 2)}
          </CodeBlock>
        </div>
      )}

      {resource.outputs && Object.keys(resource.outputs).length > 0 && (
        <div className="flex flex-col gap-2">
          <Text>Outputs</Text>
          <CodeBlock language="json">
            {JSON.stringify(resource.outputs, null, 2)}
          </CodeBlock>
        </div>
      )}
    </div>
  )
}
