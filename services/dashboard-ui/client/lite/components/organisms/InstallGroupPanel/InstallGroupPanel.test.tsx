import { useState } from 'react'
import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { IDeploymentPlanStage } from '../../../utils/deployment-plan'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { InstallGroupPanel } from './InstallGroupPanel'

afterEach(cleanup)

const STAGE: IDeploymentPlanStage = {
  id: 'grp_primary',
  name: 'Primary region',
  order: 0,
  stage: 1,
  membership: 'label_selector',
  selector: {
    match_labels: { env: 'prod' },
    not_match_labels: { tier: 'preview' },
  },
  installs: Array.from({ length: 25 }, (_, index) => ({
    id: `inst_${index}`,
    name: `install-${index + 1}`,
    cloud_platform: 'aws',
    aws_account: { region: 'us-west-2' },
    runner_status: 'active',
    sandbox_status: 'active',
    sandbox_health_status: 'healthy',
    composite_component_status: 'healthy',
  })),
  totalInstalls: 25,
  status: 'in-progress',
}

const PanelFixture = () => {
  const [offset, setOffset] = useState(0)
  return (
    <InstallGroupPanel
      stage={STAGE}
      offset={offset}
      pageSize={20}
      onOffsetChange={setOffset}
      onInstallSelect={() => {}}
    />
  )
}

const renderPanel = (panel: React.ReactElement) =>
  render(
    <MemoryRouter>
      <SurfaceStory open={({ openPanel }) => openPanel(panel)} />
    </MemoryRouter>
  )

describe('InstallGroupPanel', () => {
  test('shows the written selector and paged install membership', async () => {
    renderPanel(<PanelFixture />)

    expect(await screen.findByRole('dialog')).toBeTruthy()
    expect(screen.getByText('Installs matching these labels')).toBeTruthy()
    expect(screen.getByLabelText('not tier=preview')).toBeTruthy()
    expect(screen.getByText('install-1')).toBeTruthy()
    expect(screen.getAllByText('US West (Oregon)').length).toBe(20)
    expect(screen.getAllByLabelText('Runner active').length).toBe(20)
    expect(screen.getAllByLabelText('Sandbox healthy').length).toBe(20)
    expect(screen.getAllByLabelText('Components healthy').length).toBe(20)
    expect(screen.queryByText('install-21')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Next' }))

    expect(screen.getByText('install-21')).toBeTruthy()
    expect(screen.queryByText('install-1')).toBeNull()
  })

  test('renders approval waiting without decision controls', async () => {
    renderPanel(
      <InstallGroupPanel
        stage={{ ...STAGE, status: 'approval-awaiting' }}
        offset={0}
        pageSize={20}
        onOffsetChange={() => {}}
      />
    )

    expect(await screen.findByText('Waiting for approval')).toBeTruthy()
    expect(screen.queryByRole('button', { name: /approve/i })).toBeNull()
    expect(screen.queryByRole('button', { name: /deny/i })).toBeNull()
  })
})
