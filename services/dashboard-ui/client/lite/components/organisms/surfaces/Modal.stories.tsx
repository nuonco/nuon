import { useRef, useState } from 'react'
import { useSurfaces } from '../../../hooks/use-surfaces'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfacePlayground, SurfaceStory } from '../../__stories__/SurfaceStory'
import { Button } from '../../atoms/Button'
import { Input } from '../../atoms/Input'
import { Text } from '../../atoms/Text'
import { Modal, type TModalSize } from './Modal'

export default { title: 'lite/organisms/Modal' }

const DESCRIPTION =
  'Modals hold focused decisions and short workflows above the current page.'

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
        description: 'Required accessible title shown in the header.',
      },
      {
        name: 'children',
        type: 'ReactNode',
        description: 'Scrollable modal body.',
      },
      {
        name: 'description',
        type: 'ReactNode',
        description: 'Supporting copy grouped with the heading.',
      },
      {
        name: 'size',
        type: "'sm' | 'default' | 'lg' | 'xl' | 'full'",
        default: "'default'",
        description: 'Maximum modal width.',
      },
      {
        name: 'dismissible',
        type: 'boolean',
        default: 'true',
        description: 'Enables Escape, backdrop, and close-button dismissal.',
      },
      {
        name: 'headerActions',
        type: 'ReactNode',
        description: 'Controls displayed before the close button.',
      },
      {
        name: 'footerContent',
        type: 'ReactNode',
        description: 'Supporting content at the left of the footer.',
      },
      {
        name: 'primaryAction',
        type: 'IButton',
        description: 'Primary footer button props.',
      },
      {
        name: 'secondaryAction',
        type: 'IButton',
        description: 'Secondary footer button props.',
      },
      {
        name: 'showFooter',
        type: 'boolean',
        default: 'true',
        description: 'Controls whether the complete footer renders.',
      },
      {
        name: 'initialFocusRef',
        type: 'RefObject<HTMLElement>',
        description: 'Element focused when the modal opens.',
      },
      {
        name: 'bodyClassName',
        type: 'string',
        description: 'Additional classes for the scrollable body.',
      },
      {
        name: 'onClose',
        type: '() => void',
        description: 'Runs before the surface closes.',
      },
    ]}
  />
)

export const PropHeading = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Modal heading">
          <Text color="secondary">The heading labels the dialog.</Text>
        </Modal>
      )
    }
  />
)

export const PropDescription = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Deploy component" description={DESCRIPTION}>
          <Text>The description stays grouped with the heading.</Text>
        </Modal>
      )
    }
  />
)

const MODAL_SIZES: { label: string; size: TModalSize }[] = [
  { label: 'small', size: 'sm' },
  { label: 'default', size: 'default' },
  { label: 'large', size: 'lg' },
  { label: 'extra large', size: 'xl' },
  { label: 'full width', size: 'full' },
]

const ModalSizeControls = () => {
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
                  This modal uses size=&quot;{size}&quot;.
                </Text>
              </Modal>
            )
          }
        >
          Open {label}
        </Button>
      ))}
    </div>
  )
}

export const PropSize = () => (
  <SurfacePlayground>
    <div className="flex flex-col gap-4">
      <Text color="secondary">Open each supported size.</Text>
      <ModalSizeControls />
    </div>
  </SurfacePlayground>
)

export const PropDismissible = () => (
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
          <Text>Escape and backdrop clicks do not dismiss this modal.</Text>
        </Modal>
      )
    }
  />
)

export const PropHeaderActions = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Component details"
          headerActions={
            <Button size="sm" variant="ghost">
              View docs
            </Button>
          }
        >
          <Text color="secondary">
            Header actions render before the close control.
          </Text>
        </Modal>
      )
    }
  />
)

export const PropFooterContent = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Edit component"
          footerContent={
            <Text variant="caption" color="secondary">
              Last saved 2 minutes ago
            </Text>
          }
        >
          <Text color="secondary">
            Footer content remains separate from action buttons.
          </Text>
        </Modal>
      )
    }
  />
)

export const PropPrimaryAction = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Approve plan?"
          primaryAction={{ children: 'Approve plan', variant: 'primary' }}
        >
          <Text>The primary action appears after the cancel action.</Text>
        </Modal>
      )
    }
  />
)

export const PropSecondaryAction = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal
          heading="Save configuration"
          primaryAction={{ children: 'Save configuration', variant: 'primary' }}
          secondaryAction={{ children: 'Discard changes' }}
        >
          <Text>
            The supplied secondary action replaces the default cancel.
          </Text>
        </Modal>
      )
    }
  />
)

export const PropShowFooter = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Live operation" showFooter={false}>
          <Text color="secondary">
            showFooter=false removes the footer while retaining header
            dismissal.
          </Text>
        </Modal>
      )
    }
  />
)

const InitialFocusModal = () => {
  const inputRef = useRef<HTMLInputElement>(null)
  return (
    <Modal heading="Initial focus" initialFocusRef={inputRef}>
      <label className="flex flex-col gap-2">
        <Text variant="label">Focused when opened</Text>
        <Input ref={inputRef} placeholder="Focused input" />
      </label>
      <label className="flex flex-col gap-2">
        <Text variant="label">Second input</Text>
        <Input placeholder="Not initially focused" />
      </label>
    </Modal>
  )
}

export const PropInitialFocusRef = () => (
  <SurfaceStory open={({ openModal }) => openModal(<InitialFocusModal />)} />
)

export const PropBodyClassName = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Custom body spacing" bodyClassName="gap-8">
          <Text>First body item</Text>
          <Text>Second body item with an eight-unit gap</Text>
        </Modal>
      )
    }
  />
)

const ModalOnCloseExample = () => {
  const { openModal } = useSurfaces()
  const [closeCount, setCloseCount] = useState(0)
  return (
    <div className="flex flex-col items-start gap-4">
      <Text color="secondary">onClose called: {closeCount} times</Text>
      <Button
        onClick={() =>
          openModal(
            <Modal
              heading="Close callback"
              onClose={() => setCloseCount((count) => count + 1)}
            >
              <Text>Close this modal with any dismissal control.</Text>
            </Modal>
          )
        }
      >
        Open modal
      </Button>
    </div>
  )
}

export const PropOnClose = () => (
  <SurfacePlayground>
    <ModalOnCloseExample />
  </SurfacePlayground>
)

export const UseCaseConfirmation = () => (
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
      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-4"
      >
        <label className="flex flex-col gap-2">
          <Text variant="label">Name</Text>
          <Input ref={nameRef} placeholder="Production" />
        </label>
        <label className="flex flex-col gap-2">
          <Text variant="label">Region</Text>
          <Input placeholder="us-west-2" />
        </label>
      </form>
    </Modal>
  )
}

export const UseCaseForm = () => (
  <SurfaceStory open={({ openModal }) => openModal(<EnvironmentFormModal />)} />
)

export const UseCaseLongContent = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Activity details" size="lg">
          {Array.from({ length: 18 }, (_, index) => (
            <div key={index} className="rounded-lg border border-divider p-4">
              <Text variant="label">Activity {index + 1}</Text>
              <Text color="secondary">
                Header and footer chrome remain visible while the body scrolls.
              </Text>
            </div>
          ))}
        </Modal>
      )
    }
  />
)

export const UseCaseResponsiveWidth = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <Modal heading="Responsive modal" size="xl">
          <Text color="secondary">
            Resize the story viewport to inspect responsive gutters.
          </Text>
        </Modal>
      )
    }
  />
)
