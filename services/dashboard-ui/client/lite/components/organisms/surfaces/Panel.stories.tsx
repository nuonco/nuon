import { useState } from 'react'
import { useSurfaces } from '../../../hooks/use-surfaces'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfacePlayground, SurfaceStory } from '../../__stories__/SurfaceStory'
import { Badge } from '../../atoms/Badge'
import { Button } from '../../atoms/Button'
import { Text } from '../../atoms/Text'
import { Panel, type TPanelSize } from './Panel'

export default { title: 'lite/organisms/Panel' }

const PANEL_COPY =
  'Panels keep the current page visible while showing supporting detail or controls.'

export const Overview = () => (
  <ComponentDocs
    name="Panel"
    tier="organism"
    summary={PANEL_COPY}
    use={[
      'Use for supporting detail, inspection, and workflows that benefit from page context.',
      'Stack panels when a panel opens a related resource.',
    ]}
    avoid={[
      'Do not mount Panel directly in page content.',
      'Do not use a panel when the destination should have its own route.',
    ]}
    rules={[
      'Panels enter from the right and preserve the viewport gap.',
      'Only the topmost panel accepts focus or pointer interaction.',
      'Expandable panels return to their configured starting width.',
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
        description: 'Scrollable panel body.',
      },
      {
        name: 'defaultSize',
        type: "'default' | 'half' | 'wide' | 'full'",
        default: "'default'",
        description: 'Starting width for an uncontrolled panel.',
      },
      {
        name: 'size',
        type: "'default' | 'half' | 'wide' | 'full'",
        description: 'Controlled panel width.',
      },
      {
        name: 'expandable',
        type: 'boolean',
        default: 'true',
        description: 'Shows the expand-to-full control.',
      },
      {
        name: 'headerActions',
        type: 'ReactNode',
        description: 'Controls displayed before resize and close.',
      },
      {
        name: 'onSizeChange',
        type: '(size: TPanelSize) => void',
        description: 'Reports expand and restore size changes.',
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
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Panel heading">
          <Text color="secondary">The heading labels the dialog.</Text>
        </Panel>
      )
    }
  />
)

const PANEL_SIZES: { label: string; size: TPanelSize }[] = [
  { label: 'default', size: 'default' },
  { label: 'half width', size: 'half' },
  { label: 'wide', size: 'wide' },
  { label: 'full width', size: 'full' },
]

const DefaultSizeControls = () => {
  const { openPanel } = useSurfaces()
  return (
    <div className="flex flex-wrap gap-2">
      {PANEL_SIZES.map(({ label, size }) => (
        <Button
          key={size}
          onClick={() =>
            openPanel(
              <Panel heading={`${label} panel`} defaultSize={size}>
                <Text color="secondary">
                  This panel starts with defaultSize=&quot;{size}&quot;.
                </Text>
              </Panel>
            )
          }
        >
          Open {label}
        </Button>
      ))}
    </div>
  )
}

export const PropDefaultSize = () => (
  <SurfacePlayground>
    <div className="flex flex-col gap-4">
      <Text color="secondary">Open each uncontrolled starting size.</Text>
      <DefaultSizeControls />
    </div>
  </SurfacePlayground>
)

const ControlledSizePanel = () => {
  const [size, setSize] = useState<TPanelSize>('default')
  return (
    <Panel heading="Controlled size" size={size} expandable={false}>
      <Text color="secondary">Current size: {size}</Text>
      <div className="flex flex-wrap gap-2">
        {PANEL_SIZES.map((option) => (
          <Button key={option.size} onClick={() => setSize(option.size)}>
            Set {option.label}
          </Button>
        ))}
      </div>
    </Panel>
  )
}

export const PropSize = () => (
  <SurfaceStory open={({ openPanel }) => openPanel(<ControlledSizePanel />)} />
)

export const PropExpandable = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Fixed width" expandable={false}>
          <Text color="secondary">
            expandable=false removes the resize control.
          </Text>
        </Panel>
      )
    }
  />
)

export const PropHeaderActions = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel
          heading="Deploy details"
          headerActions={
            <>
              <Badge tone="accent">Active</Badge>
              <Button size="sm" variant="ghost">
                View logs
              </Button>
            </>
          }
        >
          <Text color="secondary">
            Header actions render before resize and close controls.
          </Text>
        </Panel>
      )
    }
  />
)

const SizeChangePanel = () => {
  const [size, setSize] = useState<TPanelSize>('half')
  const [changes, setChanges] = useState(0)
  return (
    <Panel
      heading="Size change callback"
      defaultSize="half"
      size={size}
      onSizeChange={(nextSize) => {
        setSize(nextSize)
        setChanges((count) => count + 1)
      }}
    >
      <Text color="secondary">Current size: {size}</Text>
      <Text color="secondary">onSizeChange called: {changes} times</Text>
      <Text>Use the resize control in the header.</Text>
    </Panel>
  )
}

export const PropOnSizeChange = () => (
  <SurfaceStory open={({ openPanel }) => openPanel(<SizeChangePanel />)} />
)

export const PropBodyClassName = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Custom body spacing" bodyClassName="gap-8">
          <Text>First body item</Text>
          <Text>Second body item with an eight-unit gap</Text>
        </Panel>
      )
    }
  />
)

const PanelOnCloseExample = () => {
  const { openPanel } = useSurfaces()
  const [closeCount, setCloseCount] = useState(0)
  return (
    <div className="flex flex-col items-start gap-4">
      <Text color="secondary">onClose called: {closeCount} times</Text>
      <Button
        onClick={() =>
          openPanel(
            <Panel
              heading="Close callback"
              onClose={() => setCloseCount((count) => count + 1)}
            >
              <Text>Close this panel from its header.</Text>
            </Panel>
          )
        }
      >
        Open panel
      </Button>
    </div>
  )
}

export const PropOnClose = () => (
  <SurfacePlayground>
    <PanelOnCloseExample />
  </SurfacePlayground>
)

export const UseCaseSupportingDetail = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Install details">
          <Text color="secondary">{PANEL_COPY}</Text>
        </Panel>
      )
    }
  />
)

export const UseCasePlanReview = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Terraform changes" defaultSize="wide">
          <Text color="secondary">
            A wide panel supports a dense diff without replacing the underlying
            page.
          </Text>
        </Panel>
      )
    }
  />
)

export const UseCaseLongContent = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Run history" defaultSize="half">
          {Array.from({ length: 20 }, (_, index) => (
            <div key={index} className="rounded-lg border border-divider p-4">
              <Text variant="label">Run {20 - index}</Text>
              <Text color="secondary">Completed deployment activity</Text>
            </div>
          ))}
        </Panel>
      )
    }
  />
)

export const UseCaseResponsiveWidth = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Responsive panel" defaultSize="wide">
          <Text color="secondary">
            Resize the story viewport to inspect mobile width and safe gutters.
          </Text>
        </Panel>
      )
    }
  />
)
