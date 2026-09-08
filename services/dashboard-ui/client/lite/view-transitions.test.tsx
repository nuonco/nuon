import { afterEach, beforeEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import type { ReactElement } from 'react'
import {
  createBrowserRouter,
  Link as RouterLink,
  RouterProvider,
} from 'react-router'
import { Button } from './components/atoms/Button'
import { Link } from './components/atoms/Link'
import { MenuItem } from './components/molecules/Menu'
import { NavLink } from './components/molecules/NavLink'
import { PageTransition } from './components/templates/PageTransition'
import { useNavShortcuts } from './hooks/use-nav-shortcuts'

interface IStubbedTransition {
  ready: Promise<void>
  finished: Promise<void>
  updateCallbackDone: Promise<void>
  skipTransition: () => void
}

const transitionHost = document as unknown as {
  startViewTransition?: (update: () => unknown) => IStubbedTransition
}

const testWindow = window as unknown as {
  happyDOM?: { setURL: (url: string) => void }
}

let transitions = 0

beforeEach(() => {
  transitions = 0
  testWindow.happyDOM?.setURL('http://localhost/')
  transitionHost.startViewTransition = (update) => {
    transitions += 1
    const done = Promise.resolve(update()).then(() => undefined)
    return {
      ready: done,
      finished: done,
      updateCallbackDone: done,
      skipTransition: () => {},
    }
  }
})

afterEach(() => {
  cleanup()
  delete transitionHost.startViewTransition
})

const renderTrigger = (trigger: ReactElement) => {
  const router = createBrowserRouter([
    { path: '/', element: trigger },
    { path: '/next', element: <PageTransition>Next page</PageTransition> },
  ])

  render(<RouterProvider router={router} />)
}

const clickThrough = async (
  trigger: ReactElement,
  role: 'link' | 'menuitem' = 'link'
) => {
  renderTrigger(trigger)
  fireEvent.click(screen.getByRole(role, { name: 'Go' }))
  await screen.findByText('Next page')
}

test('links navigate through a view transition', async () => {
  await clickThrough(<Link href="/next">Go</Link>)

  expect(transitions).toBe(1)
})

test('link buttons navigate through a view transition', async () => {
  await clickThrough(<Button href="/next">Go</Button>)

  expect(transitions).toBe(1)
})

test('nav links navigate through a view transition', async () => {
  await clickThrough(<NavLink href="/next" label="Go" icon="HouseIcon" />)

  expect(transitions).toBe(1)
})

test('menu items navigate through a view transition', async () => {
  await clickThrough(<MenuItem href="/next">Go</MenuItem>, 'menuitem')

  expect(transitions).toBe(1)
})

test('navigation shortcuts move through a view transition', async () => {
  const Shortcuts = () => {
    useNavShortcuts([
      { href: '/next', label: 'Go', icon: 'HouseIcon', shortcut: 'g n' },
    ])
    return <p>Start page</p>
  }

  renderTrigger(<Shortcuts />)
  fireEvent.keyDown(document, { key: 'g' })
  fireEvent.keyDown(document, { key: 'n' })
  await screen.findByText('Next page')

  expect(transitions).toBe(1)
})

test('an unopted router link keeps the immediate swap', async () => {
  await clickThrough(<RouterLink to="/next">Go</RouterLink>)

  expect(transitions).toBe(0)
})
