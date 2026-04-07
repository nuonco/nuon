export default {
  title: 'Components/Configs/HelmConfig',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { HelmValuesFilesModal, HelmValuesModal } from './HelmConfig'

export const ValuesFilesModal = () => (
  <ModalStory>
    <HelmValuesFilesModal
      valuesFiles={[
        `global:
  env: production
  region: us-east-1

replicaCount: 3
image:
  repository: my-registry/my-app
  tag: "1.2.3"`,
        `service:
  type: ClusterIP
  port: 8080`,
      ]}
    />
  </ModalStory>
)

export const ValuesModal = () => (
  <ModalStory>
    <HelmValuesModal
      values={{
        'global.env': 'production',
        'replicaCount': '3',
        'image.tag': '1.2.3',
        'service.port': '8080',
      }}
    />
  </ModalStory>
)
