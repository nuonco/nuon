import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { ConfigContext, type TRuntimeConfig } from '@/providers/config-provider'
import { ConsentProvider } from '@/providers/consent-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { ConsentToastContainer } from './ConsentToastContainer'

const config: TRuntimeConfig = {
  apiUrl: '',
  appUrl: '',
  githubAppName: '',
  isByoc: false,
  posthogKey: 'test-key',
}

afterEach(() => {
  cleanup()
  localStorage.clear()
})

const renderConsent = () =>
  render(
    <ConfigContext.Provider value={config}>
      <ConsentProvider>
        <ToastProvider>
          <ConsentToastContainer />
        </ToastProvider>
      </ConsentProvider>
    </ConfigContext.Provider>
  )

test.each(['granted', 'denied'] as const)(
  'does not prompt when %s consent is already saved',
  (status) => {
    localStorage.setItem(
      'analytics_consent',
      JSON.stringify({ status, decidedAt: '2026-01-01T00:00:00Z' })
    )

    renderConsent()

    expect(screen.queryByText('Help us improve Nuon')).toBeNull()
  }
)

test.each([
  ['Accept', 'granted'],
  ['Decline', 'denied'],
] as const)(
  'saves %s and does not prompt again after remount',
  (label, status) => {
    const view = renderConsent()
    fireEvent.click(screen.getByRole('button', { name: label }))

    const saved = JSON.parse(
      localStorage.getItem('analytics_consent') ?? 'null'
    )
    expect(saved.status).toBe(status)
    expect(saved.decidedAt).toBeTruthy()

    view.unmount()
    renderConsent()

    expect(screen.queryByText('Help us improve Nuon')).toBeNull()
  }
)
