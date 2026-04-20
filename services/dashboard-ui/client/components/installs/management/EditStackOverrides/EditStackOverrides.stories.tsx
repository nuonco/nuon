import { ModalStory } from '@/components/__stories__/helpers'
import { EditStackOverridesModal } from './EditStackOverrides'

export default {
  title: 'Installs/Management/EditStackOverrides',
}

const noop = () => {}

export const Empty = () => (
  <ModalStory label="Edit stack overrides (empty)">
    <EditStackOverridesModal
      isPending={false}
      error={null}
      currentVpcUrl=""
      currentRunnerUrl=""
      currentCustomStacks={[]}
      appDefaultVpcUrl="https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/vpc.yaml"
      appDefaultRunnerUrl="https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/runner.yaml"
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithOverrides = () => (
  <ModalStory label="Edit stack overrides (with values)">
    <EditStackOverridesModal
      isPending={false}
      error={null}
      currentVpcUrl="https://custom-bucket.s3.us-east-1.amazonaws.com/vpc-custom.yaml"
      currentRunnerUrl="https://custom-bucket.s3.us-east-1.amazonaws.com/runner-custom.yaml"
      currentCustomStacks={[
        {
          name: 'k8s_namespaces',
          template_url: 'https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/k8s-namespaces.yaml',
          index: 0,
        },
        {
          name: 'eks_access_entries',
          template_url: 'https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/eks-access.yaml',
          index: 1,
          parameters: { Namespaces: '{{.nuon.install.inputs.namespaces}}' },
        },
      ]}
      appDefaultVpcUrl="https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/vpc.yaml"
      appDefaultRunnerUrl="https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/runner.yaml"
      onSubmit={noop}
    />
  </ModalStory>
)

export const NoAppDefaults = () => (
  <ModalStory label="Edit stack overrides (no app defaults)">
    <EditStackOverridesModal
      isPending={false}
      error={null}
      currentVpcUrl=""
      currentRunnerUrl=""
      currentCustomStacks={[]}
      appDefaultVpcUrl=""
      appDefaultRunnerUrl=""
      onSubmit={noop}
    />
  </ModalStory>
)

export const Saving = () => (
  <ModalStory label="Edit stack overrides (saving)">
    <EditStackOverridesModal
      isPending={true}
      error={null}
      currentVpcUrl="https://custom-bucket.s3.us-east-1.amazonaws.com/vpc-custom.yaml"
      currentRunnerUrl=""
      currentCustomStacks={[]}
      appDefaultVpcUrl="https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/vpc.yaml"
      appDefaultRunnerUrl=""
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory label="Edit stack overrides (error)">
    <EditStackOverridesModal
      isPending={false}
      error={{ error: 'vpc_nested_template_url must be an S3 URL' }}
      currentVpcUrl="https://example.com/not-s3.yaml"
      currentRunnerUrl=""
      currentCustomStacks={[]}
      appDefaultVpcUrl="https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/vpc.yaml"
      appDefaultRunnerUrl=""
      onSubmit={noop}
    />
  </ModalStory>
)
