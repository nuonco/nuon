import { useState } from 'react'
import { useSurfaces } from '../../../hooks/use-surfaces'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfacePlayground, SurfaceStory } from '../../__stories__/SurfaceStory'
import { Badge } from '../../atoms/Badge'
import { Button } from '../../atoms/Button'
import { Text } from '../../atoms/Text'
import { Panel, type TPanelSize } from './Panel'

export default {
  title: 'lite/organisms/Panel',
}

const PANEL_COPY =
  'Panels keep the current page visible while showing supporting detail or controls.'

const PANEL_SIZES: { label: string; size: TPanelSize }[] = [
  { label: 'default', size: 'default' },
  { label: 'half width', size: 'half' },
  { label: 'wide', size: 'wide' },
  { label: 'full width', size: 'full' },
]

const PanelSizeButtons = () => {
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
                  This panel uses the {size} size variant.
                </Text>
              </Panel>
            )
          }
        >
          Open {label} panel
        </Button>
      ))}
    </div>
  )
}

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
        description: 'Accessible title shown in the panel header.',
      },
      {
        name: 'defaultSize',
        type: "'default' | 'half' | 'wide' | 'full'",
        default: "'default'",
        description: 'Starting width on medium and larger viewports.',
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
        description: 'Allows the panel to toggle to full width.',
      },
    ]}
  />
)

export const Default = () => (
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

export const Sizes = () => (
  <SurfacePlayground>
    <div className="flex flex-col gap-4">
      <Text color="secondary">Open each panel size from the same fixture.</Text>
      <PanelSizeButtons />
    </div>
  </SurfacePlayground>
)

export const HalfWidth = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Deployment plan" defaultSize="half">
          <Text color="secondary">
            Half-width panels provide more room for structured reviews.
          </Text>
        </Panel>
      )
    }
  />
)

export const Wide = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Terraform changes" defaultSize="wide">
          <Text color="secondary">
            Wide panels support dense diffs without replacing the underlying
            page.
          </Text>
        </Panel>
      )
    }
  />
)

export const Full = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Workflow trace" defaultSize="full">
          <Text color="secondary">
            Full panels retain the floating viewport inset and surface chrome.
          </Text>
        </Panel>
      )
    }
  />
)

export const FixedWidth = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Quick settings" expandable={false}>
          <Text color="secondary">
            The expand control is omitted when the workflow should remain
            compact.
          </Text>
        </Panel>
      )
    }
  />
)

export const HeaderActions = () => (
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
            Status and secondary controls can sit beside the panel title.
          </Text>
        </Panel>
      )
    }
  />
)

export const LongContent = () => (
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

const ControlledExpandedPanel = () => {
  const [size, setSize] = useState<TPanelSize>('full')
  return (
    <Panel heading="Configuration editor" size={size} onSizeChange={setSize}>
      <Text color="secondary">
        The controlled size returns this panel to its default width.
      </Text>
    </Panel>
  )
}

export const ControlledSize = () => (
  <SurfaceStory
    open={({ openPanel }) => openPanel(<ControlledExpandedPanel />)}
  />
)

export const MobileWidth = () => (
  <SurfaceStory
    open={({ openPanel }) =>
      openPanel(
        <Panel heading="Responsive panel" defaultSize="wide">
          <Text color="secondary">
            Resize the story viewport to inspect full-width mobile treatment and
            safe gutters.
          </Text>
        </Panel>
      )
    }
  />
)
