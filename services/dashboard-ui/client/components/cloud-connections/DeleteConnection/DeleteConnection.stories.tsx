import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteConnection } from './DeleteConnection'
export default { title: 'Features / Cloud connections / Delete' }
export const Confirm = () => (
  <ModalStory>
    <DeleteConnection
      name="acme-production"
      isPending={false}
      onDelete={() => {}}
    />
  </ModalStory>
)
export const Conflict = () => (
  <ModalStory>
    <DeleteConnection
      name="acme-production"
      isPending={false}
      onDelete={() => {}}
      error={{
        error: 'Conflict',
        description:
          'This connection is used by 2 installs. Delete those installs first.',
        user_error: true,
        status: 409,
      }}
    />
  </ModalStory>
)
export const GAR = () => (
  <ModalStory>
    <DeleteConnection
      name="acme-gar-images"
      isPending={false}
      onDelete={() => {}}
    />
  </ModalStory>
)
export const AssignedRepository = () => (
  <ModalStory>
    <DeleteConnection
      name="acme-ecr-images"
      orgId="org-mock-001"
      imageRepositories={1}
      isPending={false}
      onDelete={() => {}}
    />
  </ModalStory>
)
export const AssignedRepositories = () => (
  <ModalStory>
    <DeleteConnection
      name="acme-acr-images"
      orgId="org-mock-001"
      imageRepositories={3}
      isPending={false}
      onDelete={() => {}}
    />
  </ModalStory>
)
