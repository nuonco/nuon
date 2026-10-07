import { expect, spyOn, test } from 'bun:test'
import { registerSyntax as registerDashboard } from '@/lib/syntax'
import { registerSyntax as registerLite } from '@/lite/utils/syntax'

test('both apps can register syntax without registering rego twice', () => {
  const error = spyOn(console, 'error')
  registerDashboard()
  registerLite()
  const messages = error.mock.calls.map((call) => call.join(' '))
  expect(messages.some((message) => message.includes('already registered'))).toBe(false)
  error.mockRestore()
})
