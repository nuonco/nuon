export default {
  title: 'Installs/DeprovisionStack',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeprovisionStackModal } from './DeprovisionStack'

const noop = () => {}

export const AWS = () => (
  <ModalStory>
    <DeprovisionStackModal
      installName="prod-acme"
      stackType="aws-cloudformation"
      onDismiss={noop}
    />
  </ModalStory>
)

export const Azure = () => (
  <ModalStory>
    <DeprovisionStackModal
      installName="prod-acme"
      stackType="azure-bicep"
      onDismiss={noop}
    />
  </ModalStory>
)

export const GCP = () => (
  <ModalStory>
    <DeprovisionStackModal
      installName="prod-acme"
      stackType="gcp-terraform"
      onDismiss={noop}
    />
  </ModalStory>
)

export const UnknownProvider = () => (
  <ModalStory>
    <DeprovisionStackModal installName="prod-acme" onDismiss={noop} />
  </ModalStory>
)
