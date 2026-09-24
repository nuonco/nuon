import { afterEach, describe, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { moduleById } from '../../utils/modules'
import { ModuleCard } from './ModuleCard'

afterEach(cleanup)

const APPS = moduleById('apps')!
const TEAM = moduleById('team')!

describe('ModuleCard', () => {
  test('asks for the opposite value when the switch is used', () => {
    const onToggle = mock(() => {})
    render(<ModuleCard module={APPS} enabled onToggle={onToggle} />)

    fireEvent.click(screen.getByRole('switch', { name: 'Apps module' }))

    expect(onToggle).toHaveBeenCalledWith(false)
  })

  test('locks the switch when the deployment pins the module', () => {
    const onToggle = mock(() => {})
    render(
      <ModuleCard module={APPS} enabled={false} pinned onToggle={onToggle} />
    )

    const control = screen.getByRole('switch', { name: 'Apps module' })
    fireEvent.click(control)

    expect(control.hasAttribute('disabled')).toBe(true)
    expect(onToggle).not.toHaveBeenCalled()
    expect(screen.getByText('Pinned by deployment')).toBeTruthy()
  })

  test('shows readiness only for modules that are not fully built', () => {
    const view = render(<ModuleCard module={TEAM} enabled />)

    expect(screen.queryByText('Built')).toBeNull()

    view.rerender(<ModuleCard module={APPS} enabled />)

    expect(screen.getByText('Partly built')).toBeTruthy()
  })
})
