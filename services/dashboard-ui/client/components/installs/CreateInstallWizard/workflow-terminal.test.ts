import { describe, expect, test } from 'bun:test'
import { isTerminalWorkflow } from './workflow-terminal'

describe('isTerminalWorkflow', () => {
  test('recognizes every workflow exit state', () => {
    expect(isTerminalWorkflow({ status: 'success' })).toBe(true)
    expect(isTerminalWorkflow({ status: 'error' })).toBe(true)
    expect(isTerminalWorkflow({ status: 'cancelled' })).toBe(true)
    expect(isTerminalWorkflow({ status: 'in-progress' })).toBe(false)
    expect(isTerminalWorkflow({ status: 'unknown', finished: true })).toBe(true)
  })
})
