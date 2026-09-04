import { useRef } from 'react'
import { useSurfaces } from '../../../hooks/use-surfaces'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfacePlayground, SurfaceStory } from '../../__stories__/SurfaceStory'
import { Button } from '../../atoms/Button'
import { Input } from '../../atoms/Input'
import { Text } from '../../atoms/Text'
import { Modal, type TModalSize } from './Modal'

export default {
  title: 'lite/organisms/Modal',
}

const DESCRIPTION =
  'Modals hold focused decisions and short workflows above the current page.'

const MODAL_SIZES: { label: string; size: TModalSize }[] = [
  { label: 'small', size: 'sm' },
  { label: 'default', size: 'default' },
  { label: 'large', size: 'lg' },
  { label: 'extra large', size: 'xl' },
  { label: 'full width', size: 'full' },
]

const ModalSizeButtons = () => {
  const { openModal } = useSurfaces()
  return (
    <div className="flex flex-wrap gap-2">
      {MODAL_SIZES.map(({ label, size }) => (
        <Button
          key={size}
          onClick={() =>
            openModal(
              <Modal heading={`${label} modal`} size={size}>
                <Text color="secondary">
                  This modal uses the {size} size variant.
                </Text>
              </Modal>
            )
          }
        >
          Open {label} modal
        </Button>
      ))}
    </div>
  )
}

export const Overview = () => (
  <ComponentDocs
    name="Modal"
    tier="organism"
    summary={DESCRIPTION}
    use={[
      'Use for focused decisions that should interrupt the current workflow.',
      'Open through useSurfaces so stacking, focus, and exit motion remain coordinated.',
    ]}
    avoid={[
      'Do not mount Modal directly in page content.',
      'Do not use a modal for long reference content that benefits from page context.',
    ]}
    rules={[
      'Modals always render above panels.',
      'Escape and the backdrop close only the topmost surface.',
      'The body scrolls independently when content exceeds the viewport.',
    ]}
    props={[
      {
        name: 'heading',
        type: 'ReactNode',
        description: 'Accessible title shown in the modal header.',
      },
      {
        name: 'size',
        type: "'sm' | 'default' | 'lg' | 'xl' | 'full'",
        default: "'default'",
        description: 'Maximum modal width.',
      },
      {
        name: 'description',
        type: 'ReactNode',
        description: 'Optional supporting copy grouped with the heading.',
      },
      {
        name: 'dismissible',
        type: 'boolean',
        default: 'true',
        description: 'Enables Escape, backdrop, and close-button dismissal.',
      },
      {
        name: 'primaryAction',
        type: 'IButton',
        description: 'Primary footer button props.',
      },
    ]}
  />
)

export const Default = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Create install"
          description={DESCRIPTION}
          primaryAction={{ children: 'Create install', variant: 'primary' }}
        >
          <Text>Configure the new install before creating it.</Text>
        </Modal>
      )
    }
  />
)

export const Sizes = () => (
  <SurfacePlayground>
    <div className="flex flex-col gap-4">
      <Text color="secondary">Open each modal size from the same fixture.</Text>
      <ModalSizeButtons />
    </div>
  </SurfacePlayground>
)

export const SmallConfirmation = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Remove member?"
          size="sm"
          primaryAction={{ children: 'Remove member', variant: 'danger' }}
        >
          <Text>
            This person will lose access to the organization and its installs.
          </Text>
        </Modal>
      )
    }
  />
)

const EnvironmentFormModal = () => {
  const nameRef = useRef<HTMLInputElement>(null)
  return (
    <Modal
      heading="Create environment"
      initialFocusRef={nameRef}
      primaryAction={{ children: 'Create environment', variant: 'primary' }}
    >
      <label className="flex flex-col gap-2">
        <Text variant="label">Name</Text>
        <Input ref={nameRef} placeholder="Production" />
      </label>
      <label className="flex flex-col gap-2">
        <Text variant="label">Region</Text>
        <Input placeholder="us-west-2" />
      </label>
    </Modal>
  )
}

export const FormContent = () => (
  <SurfaceStory open={({ openModal }) => openModal(<EnvironmentFormModal />)} />
)

export const Large = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Review deployment plan"
          size="xl"
          primaryAction={{ children: 'Approve plan', variant: 'primary' }}
        >
          <div className="grid gap-4 md:grid-cols-2">
            <Text color="secondary">
              The wider layout leaves room for a side-by-side plan review.
            </Text>
            <Text color="secondary">
              Content remains constrained inside the floating glass surface.
            </Text>
          </div>
        </Modal>
      )
    }
  />
)

export const Full = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Full-width review" size="full">
          <Text color="secondary">
            Full modals retain a safe gutter on every viewport edge.
          </Text>
        </Modal>
      )
    }
  />
)

export const LongContent = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Activity details" size="lg">
          {Array.from({ length: 18 }, (_, index) => (
            <div key={index} className="rounded-lg border border-divider p-4">
              <Text variant="label">Activity {index + 1}</Text>
              <Text color="secondary">
                Surface headers and footers stay visible while this body
                scrolls.
              </Text>
            </div>
          ))}
        </Modal>
      )
    }
  />
)

export const HeaderAndFooterActions = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Edit component"
          headerActions={
            <Button size="sm" variant="ghost">
              View docs
            </Button>
          }
          footerContent={
            <Text variant="caption" color="secondary">
              Last saved 2 minutes ago
            </Text>
          }
          primaryAction={{ children: 'Save changes', variant: 'primary' }}
        >
          <Text color="secondary">
            Header and footer slots support secondary controls without changing
            the surface layout.
          </Text>
        </Modal>
      )
    }
  />
)

export const HeadingAndDescription = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Processing changes"
          description="The operation may take a few minutes."
          showFooter={false}
          size="sm"
        >
          <Text color="secondary">
            Progress details remain visually grouped without divider lines.
          </Text>
        </Modal>
      )
    }
  />
)

export const RequiredDecision = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Review policy"
          description="Choose an action before leaving this workflow."
          dismissible={false}
          size="sm"
          primaryAction={{ children: 'Accept policy', variant: 'primary' }}
        >
          <Text>The backdrop and Escape key do not dismiss this modal.</Text>
        </Modal>
      )
    }
  />
)

export const WithoutFooter = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Live operation" showFooter={false}>
          <Text color="secondary">
            The close control remains available in the stable header.
          </Text>
        </Modal>
      )
    }
  />
)

export const MobileWidth = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Responsive modal" size="xl">
          <Text color="secondary">
            Resize the story viewport to inspect the responsive gutters.
          </Text>
        </Modal>
      )
    }
  />
)
