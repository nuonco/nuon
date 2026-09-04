import { useEffect, useRef, useState, type ReactNode } from 'react'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { useSurfaces } from '../../../hooks/use-surfaces'
import { useToast } from '../../../hooks/use-toast'
import {
  DEFAULT_TOAST_TIMEOUT,
  ToastProvider,
  type IToastInput,
} from '../../../providers/toast-provider'
import { SurfacesProvider } from '../../../providers/surfaces-provider'
import { Button } from '../../atoms/Button'
import { Text } from '../../atoms/Text'
import { Modal } from '../surfaces/Modal'
import { Panel } from '../surfaces/Panel'
import { SurfaceHost } from '../surfaces/SurfaceHost'

export default { title: 'lite/organisms/Toast' }

export const Overview = () => (
  <ComponentDocs
    name="Toast"
    tier="organism"
    summary="A transient app-level notification presented in an expanding stack."
    use={[
      'Report asynchronous work beginning or completing.',
      'Offer one optional follow-up action that does not block the workflow.',
      'Use warning and error themes for global failures outside a form.',
    ]}
    avoid={[
      'Do not show field validation or form submission errors in a toast.',
      'Do not tint the entire surface with the status color.',
      'Do not use a toast when the user must decide before continuing.',
    ]}
    rules={[
      'Headings are plain strings that state what happened.',
      'Omit theme for an ordinary informational notification.',
      'The blue info theme is reserved for in-progress work.',
      'The neutral card, left status rail, and icon are consistent across themes.',
      'Three layers represent the compact deck; hover or focus reveals every active toast.',
      'The close control is always visible.',
    ]}
    props={[
      {
        name: 'heading',
        type: 'string',
        description: 'Required statement of what happened.',
      },
      {
        name: 'description',
        type: 'ReactNode',
        description: 'Optional entity or workflow context.',
      },
      {
        name: 'theme',
        type: "'default' | TStatusTheme",
        default: "'default'",
        description:
          'Controls the status rail, icon, and announcement priority.',
      },
      {
        name: 'action',
        type: 'IToastAction',
        description: 'One optional follow-up action.',
      },
      {
        name: 'timeout',
        type: 'number | null',
        default: '5000',
        description:
          'Dismissal delay in milliseconds; null remains persistent.',
      },
    ]}
  />
)

const StoryCanvas = ({ children }: { children: ReactNode }) => (
  <div className="min-h-[32rem] bg-surface-default p-8">{children}</div>
)

const WithProvider = ({ children }: { children: ReactNode }) => (
  <ToastProvider>
    <StoryCanvas>{children}</StoryCanvas>
  </ToastProvider>
)

const InitialToasts = ({ toasts }: { toasts: IToastInput[] }) => {
  const { addToast } = useToast()
  const added = useRef(false)

  useEffect(() => {
    if (added.current) return
    added.current = true
    for (const toast of toasts) addToast(toast)
  }, [addToast, toasts])

  return (
    <Text color="secondary">
      Hover or focus the notification stack to expand it.
    </Text>
  )
}

const InitialStory = ({ toasts }: { toasts: IToastInput[] }) => (
  <WithProvider>
    <InitialToasts toasts={toasts} />
  </WithProvider>
)

const ToastTrigger = ({
  buttonLabel,
  guidance,
  toast,
}: {
  buttonLabel: string
  guidance: string
  toast: IToastInput
}) => {
  const { addToast } = useToast()
  return (
    <div className="flex max-w-lg flex-col items-start gap-4">
      <Text color="secondary">{guidance}</Text>
      <Button variant="primary" onClick={() => addToast(toast)}>
        {buttonLabel}
      </Button>
    </div>
  )
}

export const PropHeading = () => (
  <InitialStory
    toasts={[{ heading: 'Configuration updated', timeout: null }]}
  />
)

export const PropDescription = () => (
  <InitialStory
    toasts={[
      {
        heading: 'Configuration updated',
        description: 'The app configuration now matches the repository.',
        timeout: null,
      },
    ]}
  />
)

const THEMES: IToastInput[] = [
  {
    heading: 'New activity available',
    description: 'Open the install timeline to review recent activity.',
    timeout: null,
  },
  {
    heading: 'Deploy completed',
    description: 'Payments is running on production.',
    theme: 'success',
    timeout: null,
  },
  {
    heading: 'Deploy failed',
    description: 'The runner returned an error while applying the plan.',
    theme: 'error',
    timeout: null,
  },
  {
    heading: 'Approval required',
    description: 'Review the proposed infrastructure changes.',
    theme: 'warn',
    timeout: null,
  },
  {
    heading: 'Deploying component',
    description: 'Deploying API to production. This may take a few minutes.',
    theme: 'info',
    timeout: null,
  },
  {
    heading: 'New capability available',
    description: 'Policy checks can now run before deployment.',
    theme: 'brand',
    timeout: null,
  },
  {
    heading: 'No changes detected',
    description: 'The app configuration already matches the repository.',
    theme: 'neutral',
    timeout: null,
  },
]

export const PropTheme = () => <InitialStory toasts={THEMES} />

const ActionExample = () => {
  const { addToast } = useToast()
  const [actionCount, setActionCount] = useState(0)

  return (
    <div className="flex flex-col items-start gap-4">
      <Text color="secondary">Action invoked: {actionCount} times</Text>
      <Button
        variant="primary"
        onClick={() =>
          addToast({
            heading: 'Plan ready',
            description: 'Infrastructure changes are ready for review.',
            theme: 'info',
            action: {
              label: 'View plan',
              onClick: () => setActionCount((count) => count + 1),
            },
            timeout: null,
          })
        }
      >
        Add actionable toast
      </Button>
    </div>
  )
}

export const PropAction = () => (
  <WithProvider>
    <ActionExample />
  </WithProvider>
)

export const PropTimeoutDefault = () => (
  <WithProvider>
    <ToastTrigger
      buttonLabel="Add five-second toast"
      guidance={`Click to start the default ${DEFAULT_TOAST_TIMEOUT / 1000}-second dismissal timer. Hover or focus the toast to pause it.`}
      toast={{
        heading: 'Default timer started',
        description: 'This toast dismisses after five seconds.',
        theme: 'info',
      }}
    />
  </WithProvider>
)

export const PropTimeoutCustom = () => (
  <WithProvider>
    <ToastTrigger
      buttonLabel="Add ten-second toast"
      guidance="Click to start a custom 10-second dismissal timer."
      toast={{
        heading: 'Custom timer started',
        description: 'This toast dismisses after ten seconds.',
        theme: 'info',
        timeout: 10000,
      }}
    />
  </WithProvider>
)

export const PropTimeoutPersistent = () => (
  <WithProvider>
    <ToastTrigger
      buttonLabel="Add persistent toast"
      guidance="This toast uses timeout: null and remains until dismissed."
      toast={{
        heading: 'Persistent notification',
        description: 'Use the close control to dismiss this toast.',
        timeout: null,
      }}
    />
  </WithProvider>
)

const DeckControls = () => {
  const { addToast, clearToasts } = useToast()
  const count = useRef(0)

  const add = () => {
    count.current += 1
    const index = count.current
    addToast({
      heading: `Operation ${index} updated`,
      description:
        index % 2
          ? 'A short notification for motion review.'
          : 'A longer notification that changes the measured height and exercises smooth stack reflow.',
      theme: index % 3 === 0 ? 'error' : index % 2 ? 'success' : 'info',
      timeout: null,
    })
  }

  return (
    <div className="flex flex-col items-start gap-4">
      <Text color="secondary">
        Add notifications individually or as a burst. Three layers remain
        visible while collapsed; hover or focus reveals every toast.
      </Text>
      <div className="flex flex-wrap gap-2">
        <Button variant="primary" onClick={add}>
          Add toast
        </Button>
        <Button
          onClick={() => {
            for (let index = 0; index < 6; index += 1) add()
          }}
        >
          Add six
        </Button>
        <Button onClick={clearToasts}>Clear toasts</Button>
      </div>
    </div>
  )
}

export const BehaviorStacking = () => (
  <WithProvider>
    <DeckControls />
  </WithProvider>
)

export const UseCaseAsyncStart = () => (
  <InitialStory
    toasts={[
      {
        heading: 'Deploying component',
        description:
          'Deploying API to production. This may take a few minutes.',
        theme: 'info',
        timeout: null,
      },
    ]}
  />
)

export const UseCaseStandardInfo = () => (
  <InitialStory
    toasts={[
      {
        heading: 'New activity available',
        description: 'Open the install timeline to review recent activity.',
        timeout: null,
      },
    ]}
  />
)

export const UseCaseCompletion = () => (
  <InitialStory
    toasts={[
      {
        heading: 'Plan approved',
        description: 'Approved changes for API on production.',
        theme: 'success',
        timeout: null,
      },
    ]}
  />
)

export const UseCaseFailure = () => (
  <InitialStory
    toasts={[
      {
        heading: 'Build failed',
        description: 'Unable to build the API component.',
        theme: 'error',
        timeout: null,
      },
    ]}
  />
)

const LayeredPanelBody = () => {
  const { addToast } = useToast()
  const { openModal } = useSurfaces()

  return (
    <div className="flex flex-col items-start gap-4">
      <Text color="secondary">
        Toasts remain visible above active panels and modals.
      </Text>
      <Button
        variant="primary"
        onClick={() =>
          addToast({
            heading: 'Panel action completed',
            description: 'The panel remains open behind this notification.',
            theme: 'success',
            timeout: null,
          })
        }
      >
        Add toast
      </Button>
      <Button
        onClick={() =>
          openModal(
            <Modal heading="Confirm change">
              <LayeredPanelBody />
            </Modal>
          )
        }
      >
        Open modal
      </Button>
    </div>
  )
}

const LayeringControls = () => {
  const { openPanel } = useSurfaces()
  return (
    <Button
      variant="primary"
      onClick={() =>
        openPanel(
          <Panel heading="Deployment details">
            <LayeredPanelBody />
          </Panel>
        )
      }
    >
      Open panel
    </Button>
  )
}

export const UseCaseAboveSurfaces = () => (
  <ToastProvider>
    <SurfacesProvider>
      <SurfaceHost scope="toast-story">
        <StoryCanvas>
          <LayeringControls />
        </StoryCanvas>
      </SurfaceHost>
    </SurfacesProvider>
  </ToastProvider>
)
