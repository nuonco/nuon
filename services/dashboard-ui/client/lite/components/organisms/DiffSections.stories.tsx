import { ComponentDocs } from '../__stories__/ComponentDocs'
import {
  deploymentAfter,
  deploymentBefore,
  longManifestDiff,
  terraformResourceDiff,
} from '../../lib/fixtures/diffs'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'

export default {
  title: 'lite/organisms/DiffSections',
}

export const Overview = () => (
  <ComponentDocs
    name="DiffSections"
    tier="organism"
    summary="A group of resource diffs with one set of reading controls."
    use={[
      'Present every resource change in one plan view.',
      'Let a reader expand all sections or change the comparison layout once.',
      'Keep each resource in its own bounded, virtualized scroll region.',
    ]}
    avoid={[
      'Do not add a separate view toggle to each section.',
      'Do not use it to parse a provider-specific plan. An adapter organism owns that job.',
      'Do not put unrelated page actions in its controls.',
    ]}
    rules={[
      'Unified or split applies to every section in the group. Wrapping and back to top belong to each diff toolbar.',
      'The expand and view controls share the toolbar row, and disappear when the group has no sections.',
      'Unified is the default so a plan remains readable at narrow widths.',
      'Collapsed sections do not mount a renderer or enter the worker queue.',
      'Each open section scrolls independently when its diff reaches the height limit.',
    ]}
    props={[
      {
        name: 'children',
        type: 'ReactNode',
        description: 'DiffSection children. Required.',
      },
      {
        name: 'toolbar',
        type: 'ReactNode',
        description:
          'Filter row rendered beside the expand and view controls, usually a DiffFilter.',
      },
      {
        name: 'defaultOpen',
        type: 'boolean',
        default: 'false',
        description: 'Starting state for every section.',
      },
      {
        name: 'defaultView',
        type: "'unified' | 'split'",
        default: "'unified'",
        description: 'Starting comparison layout.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-5xl p-8">
    <DiffSections defaultOpen>
      <DiffSection
        title="apps/api"
        description="Deployment · production"
        operation="update"
        before={deploymentBefore}
        after={deploymentAfter}
        language="yaml"
      />
      <DiffSection
        title="apps/worker"
        description="Deployment · production"
        operation="create"
        before=""
        after={deploymentAfter.replace('name: api', 'name: worker')}
        language="yaml"
      />
      <DiffSection
        title="aws_eks_cluster.this"
        description="Managed cluster"
        operation="update"
        {...terraformResourceDiff}
        filename={undefined}
      />
    </DiffSections>
  </div>
)

export const ManySections = () => {
  const large = longManifestDiff()

  return (
    <div className="max-w-5xl p-8">
      <DiffSections>
        {Array.from({ length: 30 }, (_, index) => {
          const largeSection = index === 12
          return (
            <DiffSection
              key={index}
              title={`resource_${String(index + 1).padStart(2, '0')}`}
              description={
                largeSection ? 'Large configuration' : 'Small configuration'
              }
              operation={index % 4 === 0 ? 'create' : 'update'}
              before={largeSection ? large.before : `replicas = ${index + 1}`}
              after={largeSection ? large.after : `replicas = ${index + 2}`}
              language={largeSection ? 'yaml' : 'terraform'}
            />
          )
        })}
      </DiffSections>
    </div>
  )
}

export const SplitView = () => (
  <div className="max-w-5xl p-8">
    <DiffSections defaultOpen defaultView="split">
      <DiffSection
        title="apps/api"
        description="Deployment · production"
        operation="update"
        before={deploymentBefore}
        after={deploymentAfter}
        language="yaml"
        defaultWrap
      />
      <DiffSection
        title="aws_eks_cluster.this"
        description="Managed cluster"
        operation="update"
        {...terraformResourceDiff}
        filename={undefined}
      />
    </DiffSections>
  </div>
)
