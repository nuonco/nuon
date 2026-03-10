import { useState } from 'react'
import { Badge, TBadgeTheme } from '@/components/common/Badge'
import { Code } from '@/components/common/Code'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { JSONViewer } from '@/components/common/JSONViewer'
import { LabeledValue } from '@/components/common/LabeledValue'
import { SearchInput } from '@/components/common/SearchInput'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import type { TTerraformState } from '@/types'

function getTerraformOutputsFromState(state: TTerraformState) {
  return state?.values?.outputs || {}
}

function getTerraformResourcesFromState(state: TTerraformState) {
  const rootModule = state?.values?.root_module
  const top = rootModule?.resources || []
  const nested =
    rootModule?.child_modules?.flatMap((m) => m.resources || []) || []
  return [...top, ...nested]
}

export const TerraformState = ({
  terraformState,
}: {
  terraformState: TTerraformState
}) => {
  const [outputSearch, setOutputSearch] = useState('')
  const [resourceSearch, setResourceSearch] = useState('')

  const tfOutputs = getTerraformOutputsFromState(terraformState)
  const tfResources = getTerraformResourcesFromState(terraformState)

  const filteredOutputEntries = Object.entries(tfOutputs).filter(([key]) =>
    key.toLowerCase().includes(outputSearch.toLowerCase())
  )

  const filteredResources = tfResources.filter(
    (r) =>
      r.address?.toLowerCase().includes(resourceSearch.toLowerCase()) ||
      r.name?.toLowerCase().includes(resourceSearch.toLowerCase())
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
                {filteredOutputEntries.length} of{' '}
                {Object.entries(tfOutputs).length}
              </Text>
            </div>
            <div className="flex flex-col gap-4">
              {filteredOutputEntries.length ? (
                filteredOutputEntries.map(([key, output]) => {
                  const val = output?.value
                  return (
                    <LabeledValue
                      className="border rounded-md p-4"
                      key={key}
                      label={key}
                    >
                      {Array.isArray(val) ? (
                        <CodeBlock language="json">
                          {JSON.stringify(val, null, 2)}
                        </CodeBlock>
                      ) : val !== null &&
                        val !== undefined &&
                        typeof val === 'object' ? (
                        <JSONViewer
                          data={val}
                          expanded={1}
                          showDataTypes={false}
                          showSize={false}
                          className="!border-0 !rounded-none"
                        />
                      ) : val !== null && val !== undefined ? (
                        <Code
                          className="w-full overflow-x-auto !text-xs !p-2"
                          variant="inline"
                        >
                          {String(val)}
                        </Code>
                      ) : (
                        <Text variant="subtext" theme="neutral">
                          —
                        </Text>
                      )}
                    </LabeledValue>
                  )
                })
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
                {filteredResources.length} of {tfResources.length}
              </Text>
            </div>
            <div className="flex flex-col gap-8">
              <div className="flex flex-col gap-4">
                {tfResources?.length ? (
                  filteredResources.length ? (
                    filteredResources.map((resource) => (
                      <Expand
                        id={resource.address}
                        key={resource?.address}
                        className="border rounded-md"
                        headerClassName="!px-4 py-2"
                        heading={<ResourceHeading resource={resource} />}
                      >
                        <ResourceDetails resource={resource} />
                      </Expand>
                    ))
                  ) : (
                    <Text className="p-4" variant="body" theme="neutral">
                      No resources match your search.
                    </Text>
                  )
                ) : (
                  <Text>No Terraform resoucres</Text>
                )}
              </div>
            </div>
          </div>
        ),
      }}
    />
  )
}

const RESOURCE_MODE_THEME: Record<string, TBadgeTheme> = {
  managed: 'info',
  data: 'brand',
  import: 'neutral',
  tainted: 'warn',
  orphaned: 'error',
}

const ResourceHeading = ({ resource }) => {
  return (
    <div className="flex flex-col gap-2 w-full">
      <div className="flex items-center justify-between">
        <span className="flex flex-col gap-2 text-left">
          <Text weight="strong">{resource?.address}</Text>
          <Text className="!flex items-center gap-6" theme="neutral">
            <Text variant="label">
              <b>Name:</b> {resource?.name}
            </Text>
            <Text variant="label">
              <b>Provider name:</b> {resource?.provider_name}
            </Text>
            {resource?.index && (
              <Text variant="label">
                <b>Index</b>: {resource?.index}
              </Text>
            )}
            <Text variant="label">
              <b>Schema version</b>: {resource?.schema_version}
            </Text>
          </Text>
        </span>
        <span className="flex flex-col gap-2">
          <Badge className="self-end" variant="code" size="sm">
            {resource?.type}
          </Badge>
          <Badge
            className="self-end"
            variant="code"
            size="sm"
            theme={RESOURCE_MODE_THEME[resource?.mode]}
          >
            {resource?.mode}
          </Badge>
        </span>
      </div>
    </div>
  )
}

function parseResourceValues(resource): any {
  return Object.entries(resource?.values).reduce((acc, [key, value]) => {
    if (!resource?.sensitive_values?.hasOwnProperty(key)) {
      acc[key] = value
    }
    return acc
  }, {})
}

const ResourceDetails = ({ resource }) => {
  const { yaml_body_parsed, manifest, ...values } =
    parseResourceValues(resource)

  return (
    <div className="p-4 border-t flex flex-col gap-6 max-h-180 overflow-auto">
      {resource?.depends_on && (
        <div className="flex flex-col gap-2">
          <Text>Depends on:</Text>
          <div className="flex items-center gap-2 flex-wrap">
            {resource?.depends_on?.map((d) => (
              <Badge variant="code" size="sm" key={d}>
                {d}
              </Badge>
            ))}
          </div>
        </div>
      )}

      <div className="flex flex-col gap-2">
        <Text>Values</Text>
        <CodeBlock language="json">{JSON.stringify(values, null, 2)}</CodeBlock>
      </div>

      {yaml_body_parsed && (
        <div className="flex flex-col gap-2">
          <Text>YAML body</Text>

          <CodeBlock language="yml">{yaml_body_parsed}</CodeBlock>
        </div>
      )}

      {manifest && (
        <div className="flex flex-col gap-2">
          <Text>Manifest</Text>

          <CodeBlock language="json">
            {JSON.stringify(JSON.parse(manifest), null, 2)}
          </CodeBlock>
        </div>
      )}
    </div>
  )
}
