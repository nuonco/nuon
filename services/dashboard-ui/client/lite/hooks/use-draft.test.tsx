import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { saveDraft } from '../utils/draft'
import { useDraft } from './use-draft'

afterEach(() => {
  cleanup()
  localStorage.clear()
})

const Harness = ({
  orgId = 'org_a',
  resourceId,
}: {
  orgId?: string
  resourceId?: string
}) => {
  const draft = useDraft<{ name: string }>({
    orgId,
    wizard: 'app',
    resourceId,
  })

  return (
    <div>
      <output data-testid="open">{String(draft.resumeOpen)}</output>
      <output data-testid="values">{JSON.stringify(draft.values ?? null)}</output>
      <output data-testid="age">{draft.updatedAt ?? ''}</output>
      <button onClick={draft.resume}>Resume</button>
      <button onClick={draft.startFresh}>Fresh</button>
      <button onClick={() => draft.save({ name: 'Saved' })}>Save</button>
      <button onClick={draft.clear}>Clear</button>
    </div>
  )
}

describe('useDraft', () => {
  test('offers a stored draft on re-entry and hydrates on resume', () => {
    saveDraft('org_a', 'app', { name: 'Payments' })
    render(<Harness />)

    expect(screen.getByTestId('open').textContent).toBe('true')
    expect(screen.getByTestId('values').textContent).toBe('null')

    fireEvent.click(screen.getByRole('button', { name: 'Resume' }))

    expect(screen.getByTestId('open').textContent).toBe('false')
    expect(screen.getByTestId('values').textContent).toBe('{"name":"Payments"}')
  })

  test('start fresh clears storage', () => {
    saveDraft('org_a', 'app', { name: 'Payments' })
    render(<Harness />)

    fireEvent.click(screen.getByRole('button', { name: 'Fresh' }))

    expect(screen.getByTestId('open').textContent).toBe('false')
    expect(screen.getByTestId('values').textContent).toBe('null')
    expect(localStorage.getItem('nuon-lite:draft:org_a:app:new')).toBeNull()
  })

  test('migrates the new draft when a resource id appears', () => {
    saveDraft('org_a', 'app', { name: 'Payments' })
    const view = render(<Harness />)

    view.rerender(<Harness resourceId="app_payments" />)

    expect(localStorage.getItem('nuon-lite:draft:org_a:app:new')).toBeNull()
    expect(
      localStorage.getItem('nuon-lite:draft:org_a:app:app_payments')
    ).toBeTruthy()
  })

  test('save writes values and clear removes them', () => {
    render(<Harness resourceId="app_payments" />)

    fireEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(screen.getByTestId('values').textContent).toBe('{"name":"Saved"}')

    fireEvent.click(screen.getByRole('button', { name: 'Clear' }))
    expect(screen.getByTestId('values').textContent).toBe('null')
    expect(
      localStorage.getItem('nuon-lite:draft:org_a:app:app_payments')
    ).toBeNull()
  })
})
