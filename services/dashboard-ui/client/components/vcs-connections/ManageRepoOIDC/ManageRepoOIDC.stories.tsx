import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ModalStory } from '@/components/__stories__/helpers'
import type { TOIDCTrustPolicy } from '@/types'
import { ManageRepoOIDCModal } from './ManageRepoOIDC'

export default {
  title: 'VCS Connections/ManageRepoOIDC',
}

const createSlot = (
  <Button variant="secondary" size="sm" className="w-fit">
    <Icon variant="PlusIcon" size={14} />
    Create trust policy
  </Button>
)

const renderDelete = () => (
  <Button variant="ghost" size="sm">
    <Icon variant="TrashIcon" size={14} />
    Delete
  </Button>
)

const policies: TOIDCTrustPolicy[] = [
  {
    id: 'oidcpolicy1',
    name: 'github-app',
    role: 'org_read_only',
    enabled: true,
    claim_conditions: { sub: 'repo:acme/app:ref:refs/heads/*' },
  },
]

export const WithPolicy = () => (
  <ModalStory>
    <ManageRepoOIDCModal
      policies={policies}
      isLoading={false}
      onToggle={() => {}}
      createSlot={createSlot}
      renderDelete={renderDelete}
    />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory>
    <ManageRepoOIDCModal
      policies={[]}
      isLoading={false}
      onToggle={() => {}}
      createSlot={createSlot}
      renderDelete={renderDelete}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ManageRepoOIDCModal
      policies={[]}
      isLoading
      onToggle={() => {}}
      createSlot={createSlot}
      renderDelete={renderDelete}
    />
  </ModalStory>
)
