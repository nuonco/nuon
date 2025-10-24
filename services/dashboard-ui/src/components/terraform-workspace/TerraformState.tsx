import { Badge, TBadgeTheme } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
// import type { } from "@/types";

function getTerraformOutputsFromState(state) {
  return state?.values?.outputs || {}
}

function getTerraformResourcessFromState(state) {
  return state?.values?.root_module?.resources || []
}

export const TerraformState = ({
  terraformState,
}: {
  terraformState: Record<string, any>
}) => {
  const tfOutputs = getTerraformOutputsFromState(terraformState)
  const tfResources = getTerraformResourcessFromState(terraformState)

  return (
    <div className="flex flex-col gap-6 pb-14">
      <Card className="flex flex-col gap-2">
        <Text weight="strong">Terraform outputs</Text>
        <KeyValueList
          values={Object?.keys(tfOutputs)?.map((key) => ({
            key,
            value: Array.isArray(tfOutputs[key]?.type)
              ? JSON.stringify(tfOutputs[key]?.value, null, 2)
              : tfOutputs[key]?.value,
            type: Array.isArray(tfOutputs[key]?.type)
              ? tfOutputs[key]?.type?.at(0)
              : tfOutputs[key]?.type,
          }))}
        />
      </Card>

      <Card className="!p-0 !gap-0">
        <div className="p-6">
          <Text weight="strong">Terraform resoucres</Text>
        </div>

        {tfResources?.length ? (
          tfResources?.map((resource) => (
            <Expand
              id={resource.address}
              key={resource?.address}
              className="border-t"
              headerClassName="px-6 py-2"
              heading={<ResourceHeading resource={resource} />}
            >
              <ResourceDetails resource={resource} />
            </Expand>
          ))
        ) : (
          <>no tf resoucres</>
        )}
      </Card>
    </div>
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
          <Text className="flex items-center gap-6" theme="neutral">
            <Text variant="subtext">
              <b>Name:</b> {resource?.name}
            </Text>
            <Text variant="subtext">
              <b>Provider name:</b> {resource?.provider_name}
            </Text>
            {resource?.index && (
              <Text variant="subtext">
                <b>Index</b>: {resource?.index}
              </Text>
            )}
            <Text variant="subtext">
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
    <div className="p-6 border-t flex flex-col gap-6">
      {resource?.depends_on && (
        <Text className="flex items-center gap-2 flex-wrap">
          <b>Depends on:</b>{' '}
          {resource?.depends_on?.map((d) => (
            <Badge variant="code" size="sm" key={d}>
              {d}
            </Badge>
          ))}
        </Text>
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
