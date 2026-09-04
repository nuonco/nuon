import { afterEach, expect, spyOn, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import {
  createContext,
  useContext,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { MemoryRouter, useLocation, useNavigate } from 'react-router'
import { useSurfaces } from '../../../hooks/use-surfaces'
import {
  panelRegistration,
  SurfacesProvider,
  type ISurfaceRegistration,
} from '../../../providers/surfaces-provider'
import { Button } from '../../atoms/Button'
import { Input } from '../../atoms/Input'
import { Text } from '../../atoms/Text'
import { Modal } from './Modal'
import { Panel } from './Panel'
import { SurfaceHost, useCurrentSurface } from './SurfaceHost'

afterEach(cleanup)

const SurfaceTestRoot = ({
  children,
  initialEntry = '/org-123',
  registrations = [],
}: {
  children: ReactNode
  initialEntry?: string
  registrations?: ISurfaceRegistration[]
}) => (
  <MemoryRouter initialEntries={[initialEntry]}>
    <SurfacesProvider>
      <SurfaceHost scope="test" registrations={registrations}>
        {children}
      </SurfaceHost>
    </SurfacesProvider>
  </MemoryRouter>
)

const StackControls = () => {
  const { openModal, openPanel } = useSurfaces()
  return (
    <div data-testid="page">
      <Button
        onClick={() => {
          openPanel(<Panel heading="First panel">First</Panel>)
          openPanel(<Panel heading="Second panel">Second</Panel>)
        }}
      >
        Open panels
      </Button>
      <Button
        onClick={() => {
          openPanel(<Panel heading="Panel">Panel</Panel>)
          openModal(<Modal heading="Modal">Modal</Modal>)
        }}
      >
        Open mixed stack
      </Button>
      <Button
        onClick={() => {
          openModal(<Modal heading="First modal">First</Modal>)
          openModal(<Modal heading="Second modal">Second</Modal>)
        }}
      >
        Open modals
      </Button>
    </div>
  )
}

test('stacks panels and closes only the topmost layer with Escape', async () => {
  render(
    <SurfaceTestRoot>
      <StackControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open panels' }))

  await waitFor(() =>
    expect(screen.getAllByRole('dialog', { hidden: true })).toHaveLength(2)
  )
  const dialogs = screen.getAllByRole('dialog', { hidden: true })
  expect(dialogs[0].getAttribute('aria-hidden')).toBe('true')
  expect(dialogs[1].hasAttribute('aria-hidden')).toBe(false)

  fireEvent.keyDown(dialogs[1], { key: 'Escape' })

  await waitFor(
    () =>
      expect(screen.getAllByRole('dialog', { hidden: true })).toHaveLength(1),
    { timeout: 500 }
  )
  expect(screen.getByRole('heading', { name: 'First panel' })).toBeTruthy()
})

test('closes only the topmost layer from the shared backdrop', async () => {
  render(
    <SurfaceTestRoot>
      <StackControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open panels' }))
  await waitFor(() =>
    expect(screen.getAllByRole('dialog', { hidden: true })).toHaveLength(2)
  )

  const overlays = document.querySelectorAll('[data-surface-overlay]')
  expect(overlays).toHaveLength(1)
  fireEvent.click(overlays[0])

  await waitFor(
    () => expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(1),
    { timeout: 500 }
  )
  expect(screen.getByRole('heading', { name: 'First panel' })).toBeTruthy()
})

test('keeps a modal above the complete panel stack', async () => {
  render(
    <SurfaceTestRoot>
      <StackControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open mixed stack' }))

  await waitFor(() =>
    expect(screen.getAllByRole('dialog', { hidden: true })).toHaveLength(2)
  )
  expect(screen.getByRole('dialog').textContent).toContain('Modal')
  expect(
    screen
      .getByRole('heading', { name: 'Panel', hidden: true })
      .closest('[role="dialog"]')
      ?.getAttribute('aria-hidden')
  ).toBe('true')
})

test('stacks multiple modals while keeping only the top layer interactive', async () => {
  render(
    <SurfaceTestRoot>
      <StackControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open modals' }))

  await waitFor(() =>
    expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(2)
  )
  const dialogs = screen.getAllByRole('dialog', { hidden: true })
  expect(dialogs[0].inert).toBe(true)
  expect(dialogs[0].getAttribute('aria-hidden')).toBe('true')
  expect(dialogs[1].inert).toBe(false)
})

const ProviderValue = createContext('')

const ContextPanel = () => (
  <Panel heading="Context panel">
    <Text>{useContext(ProviderValue)}</Text>
  </Panel>
)

const ContextControls = () => {
  const { openPanel } = useSurfaces()
  return (
    <Button onClick={() => openPanel(<ContextPanel />)}>Open context</Button>
  )
}

test('renders surfaces inside the provider context that owns their host', async () => {
  render(
    <MemoryRouter>
      <SurfacesProvider>
        <ProviderValue.Provider value="organization value">
          <SurfaceHost scope="organization">
            <ContextControls />
          </SurfaceHost>
        </ProviderValue.Provider>
      </SurfacesProvider>
    </MemoryRouter>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open context' }))

  expect(await screen.findByText('organization value')).toBeTruthy()
})

const LocationProbe = () => {
  const location = useLocation()
  return <output data-testid="location">{location.search}</output>
}

const UrlControls = () => {
  const { openPanelKey } = useSurfaces()
  const navigate = useNavigate()
  return (
    <>
      <Button onClick={() => openPanelKey('deploy:dpl-123')}>
        Open URL panel
      </Button>
      <Button onClick={() => navigate(1)}>Forward</Button>
      <LocationProbe />
    </>
  )
}

const URL_REGISTRATIONS = [
  panelRegistration('install', () => (
    <Panel heading="Install panel">Install</Panel>
  )),
  panelRegistration('deploy', ({ resourceId }) => (
    <Panel heading="Deploy panel">{resourceId}</Panel>
  )),
  panelRegistration('step', () => <Panel heading="Step panel">Step</Panel>),
]

test('preserves feature query parameters when opening and closing a URL panel', async () => {
  render(
    <SurfaceTestRoot
      initialEntry="/org-123?tab=logs"
      registrations={URL_REGISTRATIONS}
    >
      <UrlControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open URL panel' }))

  await screen.findByRole('heading', { name: 'Deploy panel' })
  expect(screen.getByTestId('location').textContent).toBe(
    '?tab=logs&panel=deploy%3Adpl-123'
  )

  fireEvent.click(screen.getByRole('button', { name: 'Close panel' }))

  await waitFor(() =>
    expect(screen.getByTestId('location').textContent).toBe('?tab=logs')
  )

  fireEvent.click(screen.getByRole('button', { name: 'Forward' }))

  expect(
    await screen.findByRole('heading', { name: 'Deploy panel' })
  ).toBeTruthy()
})

const MixedOrderingControls = () => {
  const { openPanel, openPanelKey } = useSurfaces()
  return (
    <>
      <Button
        onClick={() =>
          openPanel(<Panel heading="Imperative panel">Imperative</Panel>)
        }
      >
        Open imperative
      </Button>
      <Button onClick={() => openPanelKey('deploy:dpl-123')}>
        Open registered
      </Button>
    </>
  )
}

test('uses opening order across imperative and URL-backed panels', async () => {
  render(
    <SurfaceTestRoot registrations={URL_REGISTRATIONS}>
      <MixedOrderingControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open imperative' }))
  await screen.findByRole('heading', { name: 'Imperative panel' })
  fireEvent.click(
    screen.getByRole('button', { name: 'Open registered', hidden: true })
  )

  expect(
    await screen.findByRole('heading', { name: 'Deploy panel' })
  ).toBeTruthy()
  expect(
    screen
      .getByRole('heading', { name: 'Imperative panel', hidden: true })
      .closest('[role="dialog"]')
      ?.getAttribute('aria-hidden')
  ).toBe('true')
})

const OpenUrlStack = () => {
  const { openPanelKey } = useSurfaces()
  return (
    <>
      <Button
        onClick={() => {
          openPanelKey('install')
          openPanelKey('deploy:dpl-123')
        }}
      >
        Open URL stack
      </Button>
      <LocationProbe />
    </>
  )
}

test('serializes multiple URL opens issued in one interaction', async () => {
  render(
    <SurfaceTestRoot registrations={URL_REGISTRATIONS}>
      <OpenUrlStack />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open URL stack' }))

  await waitFor(() =>
    expect(screen.getByTestId('location').textContent).toBe(
      '?panel=install&panel=deploy%3Adpl-123'
    )
  )
  expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(2)
  expect(screen.getByRole('heading', { name: 'Deploy panel' })).toBeTruthy()

  fireEvent.click(screen.getByRole('button', { name: 'Close panel' }))
  await waitFor(() =>
    expect(screen.getByTestId('location').textContent).toBe('?panel=install')
  )
})

test('truncates a cold-loaded URL stack when its top layer closes', async () => {
  render(
    <SurfaceTestRoot
      initialEntry="/org-123?tab=logs&panel=install&panel=deploy%3Adpl-123"
      registrations={URL_REGISTRATIONS}
    >
      <LocationProbe />
    </SurfaceTestRoot>
  )

  await waitFor(() =>
    expect(screen.getAllByRole('dialog', { hidden: true })).toHaveLength(2)
  )
  fireEvent.click(screen.getByRole('button', { name: 'Close panel' }))

  await waitFor(() =>
    expect(screen.getByTestId('location').textContent).toBe(
      '?tab=logs&panel=install'
    )
  )
})

const CloseBottomPanel = () => {
  const { closeSurface, panels } = useSurfaces()
  return (
    <Button onClick={() => panels[0] && closeSurface(panels[0].id)}>
      Close bottom panel
    </Button>
  )
}

test('truncates dependent URL layers when a lower panel closes', async () => {
  render(
    <SurfaceTestRoot
      initialEntry="/org-123?tab=logs&panel=install&panel=deploy&panel=step"
      registrations={URL_REGISTRATIONS}
    >
      <CloseBottomPanel />
      <LocationProbe />
    </SurfaceTestRoot>
  )

  await waitFor(() =>
    expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(3)
  )
  fireEvent.click(
    screen.getByRole('button', { name: 'Close bottom panel', hidden: true })
  )

  await waitFor(() =>
    expect(screen.getByTestId('location').textContent).toBe('?tab=logs')
  )
})

const ReplaceControls = () => {
  const { openPanel, replaceSurface } = useSurfaces()
  const surfaceId = useRef('')
  return (
    <Button
      onClick={() => {
        surfaceId.current = openPanel(
          <Panel heading="Original panel">
            <Button
              onClick={() =>
                replaceSurface(
                  surfaceId.current,
                  <Panel heading="Replacement panel">Replacement</Panel>
                )
              }
            >
              Replace content
            </Button>
          </Panel>
        )
      }}
    >
      Open replaceable panel
    </Button>
  )
}

test('replaces content without changing surface identity or stack position', async () => {
  render(
    <SurfaceTestRoot>
      <ReplaceControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(
    screen.getByRole('button', { name: 'Open replaceable panel' })
  )
  const original = await screen.findByRole('dialog')
  fireEvent.click(screen.getByRole('button', { name: 'Replace content' }))

  expect(
    await screen.findByRole('heading', { name: 'Replacement panel' })
  ).toBeTruthy()
  expect(screen.getByRole('dialog') === original).toBe(true)
})

const HostContent = () => {
  const { openPanel } = useSurfaces()
  return (
    <Button
      onClick={() => openPanel(<Panel heading="Disposable panel">Panel</Panel>)}
    >
      Open
    </Button>
  )
}

const HostUnmountControls = () => {
  const [mounted, setMounted] = useState(true)
  return (
    <>
      <Button onClick={() => setMounted(false)}>Unmount host</Button>
      {mounted ? (
        <SurfaceHost scope="disposable">
          <HostContent />
        </SurfaceHost>
      ) : null}
    </>
  )
}

test('removes ephemeral surfaces when their host unmounts', async () => {
  render(
    <MemoryRouter>
      <SurfacesProvider>
        <HostUnmountControls />
      </SurfacesProvider>
    </MemoryRouter>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open' }))
  await screen.findByRole('dialog')
  fireEvent.click(
    screen.getByRole('button', { name: 'Unmount host', hidden: true })
  )

  expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(0)
})

const FocusControls = () => {
  const { openModal } = useSurfaces()
  return (
    <Button
      onClick={() =>
        openModal(
          <Modal heading="Focus modal">
            <Button>Inside modal</Button>
          </Modal>
        )
      }
    >
      Open focus modal
    </Button>
  )
}

test('makes the page inert and restores the opener after close', async () => {
  render(
    <SurfaceTestRoot>
      <div data-testid="page">
        <FocusControls />
      </div>
    </SurfaceTestRoot>
  )

  const opener = screen.getByRole('button', { name: 'Open focus modal' })
  opener.focus()
  fireEvent.click(opener)

  await screen.findByRole('dialog')
  expect(screen.getByTestId('page').closest('[inert]')).toBeTruthy()

  const footerClose = screen.getByRole('button', { name: 'Close' })
  footerClose.focus()
  fireEvent.keyDown(footerClose, { key: 'Tab' })
  const headerClose = screen.getByRole('button', { name: 'Close modal' })
  expect(document.activeElement).toBe(headerClose)
  fireEvent.keyDown(headerClose, { key: 'Tab', shiftKey: true })
  expect(document.activeElement).toBe(footerClose)

  fireEvent.click(headerClose)
  expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(1)

  await waitFor(
    () => {
      expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(0)
      expect(document.activeElement).toBe(opener)
      expect(screen.getByTestId('page').closest('[inert]')).toBeNull()
    },
    { timeout: 500 }
  )
})

const ModalApiControls = () => {
  const { openModal } = useSurfaces()
  const initialFocusRef = useRef<HTMLInputElement>(null)
  return (
    <>
      <Button
        onClick={() =>
          openModal(
            <Modal heading="Undivided modal">
              <Text>Modal content</Text>
            </Modal>
          )
        }
      >
        Open undivided modal
      </Button>
      <Button
        onClick={() =>
          openModal(
            <Modal heading="Required modal" dismissible={false}>
              <Text>Required content</Text>
            </Modal>
          )
        }
      >
        Open required modal
      </Button>
      <Button
        onClick={() =>
          openModal(
            <Modal heading="Focused form" initialFocusRef={initialFocusRef}>
              <Input ref={initialFocusRef} aria-label="Environment name" />
            </Modal>
          )
        }
      >
        Open focused form
      </Button>
    </>
  )
}

test('renders a modal header and footer without divider borders', async () => {
  render(
    <SurfaceTestRoot>
      <ModalApiControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open undivided modal' }))
  await screen.findByRole('dialog')

  expect(
    screen.getByRole('heading', { name: 'Undivided modal' }).closest('header')
      ?.className
  ).not.toContain('border-b')
  expect(
    screen.getByRole('button', { name: 'Close' }).closest('footer')?.className
  ).not.toContain('border-t')
})

test('supports required modals without implicit dismissal', async () => {
  render(
    <SurfaceTestRoot>
      <ModalApiControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open required modal' }))
  const dialog = await screen.findByRole('dialog')
  expect(screen.queryByRole('button', { name: 'Close modal' })).toBeNull()

  fireEvent.keyDown(dialog, { key: 'Escape' })
  fireEvent.click(document.querySelector('[data-surface-overlay]')!)

  expect(screen.getByRole('dialog') === dialog).toBe(true)
})

test('prefers an explicitly auto-focused field', async () => {
  render(
    <SurfaceTestRoot>
      <ModalApiControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Open focused form' }))

  const input = await screen.findByRole('textbox', {
    name: 'Environment name',
  })
  expect(document.activeElement === input).toBe(true)
})

const SelfClosingModal = () => {
  const { close } = useCurrentSurface()
  return (
    <Modal
      heading="Self-closing modal"
      primaryAction={{ children: 'Finish', onClick: close, variant: 'primary' }}
    >
      <Text>Feature surfaces can close after their own work completes.</Text>
    </Modal>
  )
}

const CurrentSurfaceControls = () => {
  const { openModal } = useSurfaces()
  return (
    <Button onClick={() => openModal(<SelfClosingModal />)}>
      Open self-closing modal
    </Button>
  )
}

test('exposes safe close controls to content in the current surface', async () => {
  render(
    <SurfaceTestRoot>
      <CurrentSurfaceControls />
    </SurfaceTestRoot>
  )

  fireEvent.click(
    screen.getByRole('button', { name: 'Open self-closing modal' })
  )
  await screen.findByRole('dialog')
  fireEvent.click(screen.getByRole('button', { name: 'Finish' }))

  await waitFor(
    () => expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(0),
    { timeout: 500 }
  )
})

test('warns and declines to resolve duplicate active registrations', async () => {
  const nodeEnv = process.env.NODE_ENV
  process.env.NODE_ENV = 'development'
  const warning = spyOn(console, 'warn').mockImplementation(() => {})
  const duplicate = panelRegistration('deploy', () => (
    <Panel heading="Duplicate">Duplicate</Panel>
  ))

  render(
    <MemoryRouter initialEntries={['/org-123?panel=deploy']}>
      <SurfacesProvider>
        <SurfaceHost scope="first" registrations={[duplicate]}>
          First
        </SurfaceHost>
        <SurfaceHost scope="second" registrations={[duplicate]}>
          Second
        </SurfaceHost>
      </SurfacesProvider>
    </MemoryRouter>
  )

  await waitFor(() => expect(warning).toHaveBeenCalled())
  await waitFor(
    () => expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(0),
    { timeout: 500 }
  )

  warning.mockRestore()
  process.env.NODE_ENV = nodeEnv
})

test('leaves unresolved URL values in place without rendering a surface', async () => {
  render(
    <SurfaceTestRoot initialEntry="/org-123?tab=logs&panel=unknown">
      <LocationProbe />
    </SurfaceTestRoot>
  )

  await new Promise((resolve) => setTimeout(resolve, 50))
  expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(0)
  expect(screen.getByTestId('location').textContent).toBe(
    '?tab=logs&panel=unknown'
  )
})
