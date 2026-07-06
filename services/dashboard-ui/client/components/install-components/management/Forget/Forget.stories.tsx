export default {
  title: 'Install Components/Forget',
}

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ModalStory } from '@/components/__stories__/helpers'
import { ForgetComponentModal } from './Forget'

const noop = () => {}

const teardownAction = (
  <Button variant="secondary" onClick={noop}>
    <Icon variant="CloudArrowDownIcon" />
    Teardown component
  </Button>
)

export const Ready = () => (
  <ModalStory>
    <ForgetComponentModal
      componentName="web-server"
      isLoading={false}
      error={null}
      onConfirm={noop}
      isTornDown
      isInConfig={false}
    />
  </ModalStory>
)

export const NeedsTeardown = () => (
  <ModalStory>
    <ForgetComponentModal
      componentName="web-server"
      isLoading={false}
      error={null}
      onConfirm={noop}
      isTornDown={false}
      isInConfig
      teardownAction={teardownAction}
    />
  </ModalStory>
)

export const StillInConfig = () => (
  <ModalStory>
    <ForgetComponentModal
      componentName="web-server"
      isLoading={false}
      error={null}
      onConfirm={noop}
      isTornDown
      isInConfig
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ForgetComponentModal
      componentName="web-server"
      isLoading={true}
      error={null}
      onConfirm={noop}
      isTornDown
      isInConfig={false}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ForgetComponentModal
      componentName="web-server"
      isLoading={false}
      error={{ error: 'Component still has active deploys' }}
      onConfirm={noop}
      isTornDown
      isInConfig={false}
    />
  </ModalStory>
)
