import { createContext, useContext, useEffect, useRef } from 'react'
import { useSurfaces } from '../../../hooks/use-surfaces'
import {
  modalRegistration,
  panelRegistration,
  SurfacesProvider,
} from '../../../providers/surfaces-provider'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory, UrlSurfaceStory } from '../../__stories__/SurfaceStory'
import { Button } from '../../atoms/Button'
import { Input } from '../../atoms/Input'
import { Text } from '../../atoms/Text'
import { Modal } from './Modal'
import { Panel } from './Panel'
import { SurfaceHost } from './SurfaceHost'

export default {
  title: 'lite/organisms/Surface stacks',
}

export const Overview = () => (
  <ComponentDocs
    name="Surface stacks"
    tier="organism"
    summary="A global coordinator and context-local hosts support ordered, URL-addressable panel and modal stacks."
    use={[
      'Mount one coordinator at the Lite route root.',
      'Mount a host inside each data-provider boundary that owns surface registrations.',
    ]}
    avoid={[
      'Do not add another coordinator for nested routes.',
      'Do not render API-aware surfaces outside their provider host.',
    ]}
    rules={[
      'Panels preserve their opening order and expose covered layers.',
      'Modals always appear above the complete panel stack.',
      'URL registrations resolve in the host where they were declared.',
    ]}
  />
)

export const StackedPanels = () => (
  <SurfaceStory
    open={({ openPanel }) => {
      openPanel(
        <Panel heading="Install details">
          <Text color="secondary">The first panel remains mounted below.</Text>
          <Input defaultValue="Preserved install state" />
        </Panel>
      )
      openPanel(
        <Panel heading="Deploy details">
          <Text color="secondary">
            Closing this panel reveals the install details beneath it.
          </Text>
          <Input defaultValue="Independent deploy state" />
        </Panel>
      )
    }}
  />
)

export const ThreePanelStack = () => (
  <SurfaceStory
    open={({ openPanel }) => {
      openPanel(
        <Panel heading="Application">
          <Text color="secondary">First layer</Text>
        </Panel>
      )
      openPanel(
        <Panel heading="Install">
          <Text color="secondary">Second layer</Text>
        </Panel>
      )
      openPanel(
        <Panel heading="Deploy">
          <Text color="secondary">Third layer</Text>
        </Panel>
      )
    }}
  />
)

export const ModalAbovePanel = () => (
  <SurfaceStory
    open={({ openModal, openPanel }) => {
      openPanel(
        <Panel heading="Install settings" defaultSize="half">
          <Text color="secondary">
            The panel remains mounted while the modal has focus.
          </Text>
        </Panel>
      )
      openModal(
        <Modal
          heading="Save settings?"
          size="sm"
          primaryAction={{ children: 'Save settings', variant: 'primary' }}
        >
          <Text>Review this change before saving it.</Text>
        </Modal>
      )
    }}
  />
)

export const ModalAboveTwoPanels = () => (
  <SurfaceStory
    open={({ openModal, openPanel }) => {
      openPanel(<Panel heading="Application">Application panel</Panel>)
      openPanel(<Panel heading="Install">Install panel</Panel>)
      openModal(
        <Modal heading="Confirm deploy" size="sm">
          <Text>The modal owns focus above both panel layers.</Text>
        </Modal>
      )
    }}
  />
)

export const StackedModals = () => (
  <SurfaceStory
    open={({ openModal }) => {
      openModal(
        <Modal heading="Configure deploy">
          <Text color="secondary">The first modal retains its state.</Text>
        </Modal>
      )
      openModal(
        <Modal heading="Select component" size="sm">
          <Text color="secondary">
            Closing this modal reveals the first modal.
          </Text>
        </Modal>
      )
    }}
  />
)

const DrillInPanel = () => {
  const { openPanel } = useSurfaces()
  return (
    <Panel
      heading="Install details"
      headerActions={
        <Button
          size="sm"
          onClick={() =>
            openPanel(
              <Panel heading="Deploy details">
                <Text color="secondary">
                  This panel was opened from another panel.
                </Text>
              </Panel>
            )
          }
        >
          View deploy
        </Button>
      }
    >
      <Text color="secondary">
        Use the header action to add another layer to the stack.
      </Text>
    </Panel>
  )
}

export const DrillIn = () => (
  <SurfaceStory open={({ openPanel }) => openPanel(<DrillInPanel />)} />
)

const OrgNameContext = createContext('')

const ContextPanel = () => {
  const orgName = useContext(OrgNameContext)
  return (
    <Panel heading="Provider context">
      <Text>
        This panel can read <strong>{orgName}</strong> from a provider mounted
        above its host.
      </Text>
    </Panel>
  )
}

const ContextLauncher = () => {
  const { openPanel } = useSurfaces()
  const opened = useRef(false)

  useEffect(() => {
    if (opened.current) return
    opened.current = true
    openPanel(<ContextPanel />)
  }, [openPanel])

  return <Text color="secondary">Organization-scoped page content</Text>
}

export const ProviderContext = () => (
  <SurfacesProvider>
    <OrgNameContext.Provider value="Acme production">
      <SurfaceHost scope="organization">
        <div className="min-h-screen p-8">
          <ContextLauncher />
        </div>
      </SurfaceHost>
    </OrgNameContext.Provider>
  </SurfacesProvider>
)

const URL_REGISTRATIONS = [
  panelRegistration('install', ({ resourceId }) => (
    <Panel heading="URL-addressable install">
      <Text color="secondary">Resource ID: {resourceId}</Text>
    </Panel>
  )),
  panelRegistration('deploy', ({ resourceId }) => (
    <Panel heading="URL-addressable deploy">
      <Text color="secondary">Resource ID: {resourceId}</Text>
    </Panel>
  )),
  modalRegistration('approve-deploy', ({ resourceId }) => (
    <Modal heading="Approve URL-addressable deploy" size="sm">
      <Text color="secondary">Resource ID: {resourceId}</Text>
    </Modal>
  )),
]

export const UrlAddressablePanel = () => (
  <SurfaceStory
    registrations={URL_REGISTRATIONS}
    open={({ openPanelKey }) => openPanelKey('deploy:dpl-example')}
  />
)

export const ColdLoadedPanelStack = () => (
  <UrlSurfaceStory
    registrations={URL_REGISTRATIONS}
    search="?panel=install%3Ains-example&panel=deploy%3Adpl-example"
  />
)

export const ColdLoadedModal = () => (
  <UrlSurfaceStory
    registrations={URL_REGISTRATIONS}
    search="?modal=approve-deploy%3Adpl-example"
  />
)

export const ColdLoadedMixedStack = () => (
  <UrlSurfaceStory
    registrations={URL_REGISTRATIONS}
    search="?panel=install%3Ains-example&panel=deploy%3Adpl-example&modal=approve-deploy%3Adpl-example"
  />
)
