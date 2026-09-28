import { afterEach, expect, test } from 'bun:test'
import { openSupportChat } from './pylon-chat'

const g = globalThis as unknown as { window?: Record<string, unknown> }
const realWindow = g.window

afterEach(() => {
  g.window = realWindow
})

test('opens the Pylon chat with the message when the widget is loaded', () => {
  const pylonCalls: unknown[][] = []
  const opened: unknown[][] = []
  g.window = {
    Pylon: (...args: unknown[]) => pylonCalls.push(args),
    open: (...args: unknown[]) => opened.push(args),
  }

  openSupportChat('help me')

  expect(pylonCalls).toEqual([['showNewMessage', 'help me']])
  expect(opened).toEqual([])
})

test('falls back to the demo request form without the widget', () => {
  const opened: unknown[][] = []
  g.window = { open: (...args: unknown[]) => opened.push(args) }

  openSupportChat('help me')

  expect(opened).toEqual([['https://nuon.co/demo-request', '_blank', 'noopener,noreferrer']])
})
