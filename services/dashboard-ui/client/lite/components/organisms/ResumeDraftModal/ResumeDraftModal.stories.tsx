import { DateTime } from 'luxon'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { ResumeDraftModal } from './ResumeDraftModal'

export default { title: 'lite/organisms/ResumeDraftModal' }

const updatedAt = DateTime.now().minus({ hours: 2 }).toISO() ?? undefined

export const Overview = () => (
  <ComponentDocs
    name="ResumeDraftModal"
    tier="organism"
    summary="The re-entry prompt that offers a stored draft back or discards it."
    use={[
      'Open it when a wizard or long form finds stored field values for this resource.',
    ]}
    avoid={[
      'Do not use it for short modal forms.',
      'Do not store a step index or resource id as the only pointer to in-progress work.',
    ]}
    rules={[
      'Heading and primary action are Resume draft; the secondary is Start fresh.',
      'Draft age renders through Time with format relative.',
      'The prompt stays until the user chooses; it is not dismissible.',
    ]}
    props={[
      {
        name: 'updatedAt',
        type: 'string',
        description: 'When the stored values were last written.',
      },
      {
        name: 'onResume',
        type: '() => void',
        description: 'Hydrates the form from the stored values.',
      },
      {
        name: 'onStartFresh',
        type: '() => void',
        description: 'Clears the stored values and starts empty.',
      },
    ]}
  />
)

export const Default = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <ResumeDraftModal
          updatedAt={updatedAt}
          onResume={() => {}}
          onStartFresh={() => {}}
        />
      )
    }
  />
)
