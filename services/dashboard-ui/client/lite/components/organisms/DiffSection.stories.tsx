import { ComponentDocs } from '../__stories__/ComponentDocs'
import {
  deploymentAfter,
  deploymentBefore,
  terraformResourceDiff,
} from '../../lib/fixtures/diffs'
import { DiffSection } from './DiffSection'

export default {
  title: 'lite/organisms/DiffSection',
}

export const Overview = () => (
  <ComponentDocs
    name="DiffSection"
    tier="organism"
    summary="A resource heading that mounts its diff only while open."
    use={[
      'Give one planned resource change a name, description and operation.',
      'Keep expensive tokenization out of collapsed plan sections.',
      'Use it inside DiffSections when several resources share controls.',
    ]}
    avoid={[
      'Do not put several resource changes in one section.',
      'Do not keep state inside the diff that must survive closing the section.',
      'Do not add a second filename header unless it carries information the section title does not.',
    ]}
    rules={[
      'operation colours the left rail and the header tint. The header itself shows the added and removed line counts.',
      'Closing the section unmounts the renderer after the transition.',
      'Inside DiffSections, unified or split view comes from the group.',
    ]}
    props={[
      {
        name: 'title',
        type: 'ReactNode',
        description: 'Resource name. Required.',
      },
      {
        name: 'description',
        type: 'ReactNode',
        description: 'Resource type, namespace or other secondary context.',
      },
      {
        name: 'operation',
        type: "'create' | 'update' | 'replace' | 'delete' | 'read' | 'no-op'",
        description:
          'The plan operation. Adapters normalize provider vocabulary onto it. Required.',
      },
      {
        name: 'before',
        type: 'string',
        description: 'Previous source. Required.',
      },
      {
        name: 'after',
        type: 'string',
        description: 'Proposed source. Required.',
      },
      {
        name: 'defaultOpen',
        type: 'boolean',
        default: 'false',
        description: 'Mount the diff initially.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-4xl p-8">
    <DiffSection
      defaultOpen
      title="apps/api"
      description="Deployment · production"
      operation="update"
      before={deploymentBefore}
      after={deploymentAfter}
      language="yaml"
    />
  </div>
)

export const Terraform = () => (
  <div className="max-w-4xl p-8">
    <DiffSection
      defaultOpen
      title="aws_eks_cluster.this"
      description="Managed cluster"
      operation="update"
      {...terraformResourceDiff}
      filename={undefined}
    />
  </div>
)

export const Collapsed = () => (
  <div className="max-w-4xl p-8">
    <DiffSection
      title="apps/api"
      description="Deployment · production"
      operation="update"
      before={deploymentBefore}
      after={deploymentAfter}
      language="yaml"
    />
  </div>
)

export const Operations = () => (
  <div className="flex max-w-4xl flex-col gap-1 p-8">
    {(['create', 'update', 'replace', 'delete', 'read'] as const).map(
      (operation) => (
        <DiffSection
          key={operation}
          title={`apps/${operation}`}
          description="Deployment · production"
          operation={operation}
          before={operation === 'create' ? '' : deploymentBefore}
          after={operation === 'delete' ? '' : deploymentAfter}
          language="yaml"
        />
      )
    )}
  </div>
)
