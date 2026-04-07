export default {
  title: 'Components/Configs/TerraformConfig',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { TerraformVariablesFilesModal, TerraformVariablesModal } from './TerraformConfig'

export const VariablesFilesModal = () => (
  <ModalStory>
    <TerraformVariablesFilesModal
      variablesFiles={[
        `variable "instance_type" {
  default = "t3.small"
}

variable "region" {
  default = "us-east-1"
}`,
        `variable "min_nodes" {
  default = 2
}`,
      ]}
    />
  </ModalStory>
)

export const VariablesModal = () => (
  <ModalStory>
    <TerraformVariablesModal
      variables={{
        instance_type: 't3.small',
        region: 'us-east-1',
        min_nodes: '2',
        max_nodes: '10',
      }}
    />
  </ModalStory>
)
