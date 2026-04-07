export default {
  title: 'VCS Connections/ConnectGithub',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ConnectGithubModal } from './ConnectGithub'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ConnectGithubModal
      githubAppName="nuon-github-app"
      orgId="org-123"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ConnectGithubModal
      githubAppName="nuon-github-app"
      orgId="org-123"
      isPending={false}
      error={{ error: 'Invalid GitHub install ID' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ConnectGithubModal
      githubAppName="nuon-github-app"
      orgId="org-123"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)
