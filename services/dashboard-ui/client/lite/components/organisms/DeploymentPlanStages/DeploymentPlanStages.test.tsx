import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import type { IDeploymentPlanStage } from '../../../utils/deployment-plan'
import { DeploymentPlanStages } from './DeploymentPlanStages'

afterEach(cleanup)

const stage = (
  id: string,
  name: string,
  order: number
): IDeploymentPlanStage => ({
  id,
  name,
  order,
  stage: order + 1,
  membership: 'install_ids',
  installs: [],
  totalInstalls: 0,
})

describe('DeploymentPlanStages', () => {
  test('renders the branch and supplied stage order', () => {
    render(
      <DeploymentPlanStages
        branch={{
          id: 'br_main',
          name: 'main',
          latest_run: {
            status: 'in-progress',
            head_sha: 'a1b2c3d4e5f6',
          },
        }}
        stages={[
          stage('grp_first', 'First group', 0),
          stage('grp_second', 'Second group', 1),
        ]}
      />
    )

    expect(screen.getByText('main')).toBeTruthy()
    expect(screen.getByText('a1b2c3d')).toBeTruthy()
    expect(screen.getByText('In progress')).toBeTruthy()
    expect(
      screen
        .getAllByRole('heading', { level: 3 })
        .map((heading) => heading.textContent)
    ).toEqual(['First group', 'Second group'])
  })

  test('distinguishes empty and failed states', () => {
    const view = render(<DeploymentPlanStages stages={[]} />)

    expect(screen.getByText('No deployment plan configured')).toBeTruthy()

    view.rerender(<DeploymentPlanStages error={new Error('failed')} />)

    expect(screen.getByText('Deployment plan failed to load')).toBeTruthy()
    expect(screen.queryByText('No deployment plan configured')).toBeNull()
  })

  test('renders real stage cards while loading', () => {
    render(<DeploymentPlanStages loading />)

    expect(screen.getAllByText('Stage')).toHaveLength(3)
    expect(screen.getAllByRole('article')).toHaveLength(3)
  })

  test('shows a normal no-runs branch state', () => {
    render(
      <DeploymentPlanStages
        branch={{ id: 'br_main', name: 'main' }}
        stages={[stage('grp_first', 'First group', 0)]}
      />
    )

    expect(screen.getByText('No rollouts yet')).toBeTruthy()
    expect(screen.queryByRole('status')).toBeNull()
  })
})
