import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { InstallSummaryPanel } from './InstallSummaryPanel'

afterEach(cleanup)

const renderPanel = (panel: React.ReactElement) =>
  render(
    <MemoryRouter>
      <SurfaceStory open={({ openPanel }) => openPanel(panel)} />
    </MemoryRouter>
  )

describe('InstallSummaryPanel', () => {
  test('shows a compact summary and link to full detail', async () => {
    renderPanel(
      <InstallSummaryPanel
        install={{
          id: 'inst_alpha',
          name: 'alpha',
          cloud_platform: 'aws',
          aws_account: { region: 'us-west-2' },
          runner_status: 'active',
          sandbox_status: 'active',
          sandbox_health_status: 'healthy',
          composite_component_status: 'deploying',
          labels: { env: 'prod' },
        }}
        installHref="/org_example/installs/inst_alpha"
      />
    )

    expect(await screen.findByRole('dialog')).toBeTruthy()
    expect(screen.getByRole('heading', { name: 'alpha' })).toBeTruthy()
    expect(screen.getByText('Amazon Web Services')).toBeTruthy()
    expect(screen.getByText('US West (Oregon)')).toBeTruthy()
    expect(screen.getByText('Runner active')).toBeTruthy()
    expect(screen.getByText('Sandbox healthy')).toBeTruthy()
    expect(screen.getByText('Components deploying')).toBeTruthy()
    expect(screen.getByText('prod')).toBeTruthy()
    expect(screen.getByRole('link', { name: 'View install' })).toHaveAttribute(
      'href',
      '/org_example/installs/inst_alpha'
    )
  })

  test('keeps failure inside the panel shell', async () => {
    renderPanel(<InstallSummaryPanel error={new Error('failed')} />)

    expect(await screen.findByRole('dialog')).toBeTruthy()
    expect(screen.getByText('Install failed to load')).toBeTruthy()
  })
})
