import { ComponentDocs } from '../__stories__/ComponentDocs'
import {
  deploymentAfter,
  deploymentBefore,
  longManifestDiff,
  terraformResourceDiff,
} from '../../lib/fixtures/diffs'
import { Diff } from './Diff'

export default {
  title: 'lite/molecules/Diff',
}

export const Overview = () => (
  <ComponentDocs
    name="Diff"
    tier="molecule"
    summary="A searchable before-and-after view that keeps syntax highlighting intact."
    use={[
      'Show the exact content changed by a plan.',
      'Compare generated manifests, configuration files and Terraform resource values.',
      'Keep a large change bounded and searchable without mounting every row.',
    ]}
    avoid={[
      'Do not add operation markers to before or after. The gutter owns them.',
      'Do not render Terraform objects directly. Pass them through terraformDiff first.',
      'Do not put several resources into one Diff. Give every resource its own section.',
    ]}
    rules={[
      'before and after are plain source strings. The renderer computes the diff.',
      'Unified is the default because it remains readable in a narrow section.',
      'Search covers both sides and expands collapsed context when it jumps to a match.',
      'Added and removed rows tint the background without replacing syntax colours.',
    ]}
    props={[
      {
        name: 'before',
        type: 'string',
        description: 'The previous source. Required.',
      },
      {
        name: 'after',
        type: 'string',
        description: 'The proposed source. Required.',
      },
      {
        name: 'language',
        type: 'string',
        description: 'Syntax language or supported alias.',
      },
      {
        name: 'filename',
        type: 'string',
        description: 'Shows the renderer header when supplied.',
      },
      {
        name: 'view',
        type: "'unified' | 'split'",
        default: "'unified'",
        description: 'How the two sides are arranged.',
      },
      {
        name: 'defaultWrap',
        type: 'boolean',
        default: 'false',
        description: 'Starting state of the toolbar wrap toggle.',
      },
      {
        name: 'lineNumbers',
        type: 'boolean',
        default: 'true',
        description: 'Show source line numbers.',
      },
      {
        name: 'search',
        type: 'boolean',
        default: 'true',
        description: 'Show find-in-diff controls.',
      },
      {
        name: 'maxHeight',
        type: 'number',
        default: '640',
        description: 'Maximum height of the virtualized scroll region.',
      },
    ]}
    sections={[
      {
        heading: 'One rendering path',
        body: 'Text plans feed before and after directly. Terraform plans serialize their structured values into deterministic HCL-like text first. From that point on the renderer, collapsing, searching and colours are identical.',
      },
      {
        heading: 'Copying',
        body: 'The plus and minus indicators are generated in a non-selectable gutter rather than stored in the source. Selecting and copying a change therefore copies code, not diff markers or line numbers.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-4xl p-8">
    <Diff
      before={deploymentBefore}
      after={deploymentAfter}
      language="yaml"
      filename="deployment.yaml"
    />
  </div>
)

export const Split = () => (
  <div className="max-w-6xl p-8">
    <Diff
      before={deploymentBefore}
      after={deploymentAfter}
      language="yaml"
      filename="deployment.yaml"
      view="split"
    />
  </div>
)

export const TerraformValues = () => (
  <div className="max-w-4xl p-8">
    <Diff {...terraformResourceDiff} />
  </div>
)

export const CollapsedHunks = () => {
  const diff = longManifestDiff()
  return (
    <div className="max-w-4xl p-8">
      <Diff {...diff} language="yaml" filename="config-map.yaml" />
    </div>
  )
}

export const WithoutHeader = () => (
  <div className="max-w-4xl p-8">
    <Diff before={deploymentBefore} after={deploymentAfter} language="yaml" />
  </div>
)
