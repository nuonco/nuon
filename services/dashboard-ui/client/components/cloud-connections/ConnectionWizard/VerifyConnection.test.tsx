import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { VerifyConnection } from './VerifyConnection'
import { acrConnection } from '../__stories__/fixtures'

afterEach(cleanup)

test('ACR failures link to the failed stage even when token errors mention Entra identity', () => {
  for (const scenario of [
    {
      message:
        'ACR token exchange could not be checked. Review the registry login server and Microsoft Entra identity, then retry.',
      heading: 'ACR token exchange failed',
      step: 3,
    },
    {
      message:
        "Microsoft Entra rejected this connection's identity. Review the existing application/service principal and exact issuer, subject, and api://AzureADTokenExchange audience in step 2. Issuer discovery and JWKS must be public.",
      heading: 'Entra federation trust check failed',
      step: 2,
    },
  ]) {
    render(
      <MemoryRouter>
        <VerifyConnection
          connection={{
            ...acrConnection,
            status: 'error',
            status_message: scenario.message,
          }}
          isVerifying={false}
          onVerify={() => {}}
          setupHref="/setup"
          detailHref="/detail"
        />
      </MemoryRouter>
    )
    expect(screen.getByText(scenario.heading)).toBeTruthy()
    expect(
      screen.getByRole('link', { name: `View runbook step ${scenario.step}` })
    ).toHaveAttribute('href', `/setup#runbook-step-${scenario.step}`)
    cleanup()
  }
})
