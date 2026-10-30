export default {
  title: 'Features / VCS connections / Connect GitHub',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ConnectGithubModal } from './ConnectGithub'

export const Default = () => (
  <ModalStory>
    <ConnectGithubModal githubAppName="nuon-github-app" orgId="org-123" />
  </ModalStory>
)
