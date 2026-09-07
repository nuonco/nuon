import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Code } from './Code'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Code',
}

export const Overview = () => (
  <ComponentDocs
    name="Code"
    tier="atom"
    summary="A fragment of code inside a sentence."
    use={[
      'Name a command, a flag, a field or an environment variable in running text.',
      'Show a short literal value where the monospace matters.',
    ]}
    avoid={[
      'Do not use it for more than a line. That is a CodeBlock.',
      'Do not use it as decoration for something that is not code.',
    ]}
    rules={[
      'Inline code is never syntax highlighted. A three-word fragment gains nothing from colour and would cost a grammar to render.',
      'It sizes itself from the text around it, so it works in a caption or a heading without being told which.',
    ]}
    props={[
      { name: 'loading', type: 'boolean', description: 'Renders a skeleton at the same size.' },
      { name: 'loadingWidth', type: 'number', description: 'Skeleton width in ch.' },
    ]}
  />
)

export const InProse = () => (
  <div className="flex max-w-lg flex-col gap-4 p-8">
    <Text as="p">
      Run <Code>nuon apps sync</Code> to push the config, then check the build
      with <Code>nuon builds list</Code>.
    </Text>
    <Text as="p" variant="caption" color="secondary">
      The runner reads <Code>NUON_API_TOKEN</Code> from the environment.
    </Text>
  </div>
)

export const Loading = () => (
  <div className="flex max-w-lg flex-col gap-2 p-8">
    <Text as="p">
      Run <Code loading loadingWidth={18} /> to push the config.
    </Text>
  </div>
)
