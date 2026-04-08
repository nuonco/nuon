export default {
  title: 'Apps/CreateInstall',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateInstallModal, FormSkeleton } from './CreateInstall'

export const Default = () => (
  <ModalStory>
    <CreateInstallModal
      isLoading={false}
      hasError={false}
      config={undefined}
      configs={[]}
      isSubmitting={false}
      onFormSubmit={() => {}}
      appId="app-1"
      platform="aws"
      onSubmitAction={async () => {}}
      onCancel={() => {}}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <CreateInstallModal
      isLoading={true}
      hasError={false}
      config={undefined}
      configs={[]}
      isSubmitting={false}
      onFormSubmit={() => {}}
      appId="app-1"
      platform="aws"
      onSubmitAction={async () => {}}
      onCancel={() => {}}
    />
  </ModalStory>
)

export const Error = () => (
  <ModalStory>
    <CreateInstallModal
      isLoading={false}
      hasError={true}
      configsError={{ error: 'Unable to load app configuration' } as any}
      config={undefined}
      configs={[]}
      isSubmitting={false}
      onFormSubmit={() => {}}
      appId="app-1"
      platform="aws"
      onSubmitAction={async () => {}}
      onCancel={() => {}}
    />
  </ModalStory>
)

export const FormSkeletonStory = () => <FormSkeleton />
