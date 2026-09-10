import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { IDeploymentPlanStage } from '../../../utils/deployment-plan'
import {
  InstallGroupCard,
  MAX_INLINE_INSTALLS,
} from './InstallGroupCard'

afterEach(cleanup)

const stage = (
  fields: Partial<IDeploymentPlanStage> = {}
): IDeploymentPlanStage => ({
  id: 'grp_core',
  name: 'Core installs',
  order: 0,
  stage: 1,
  membership: 'install_ids',
  installs: [{ id: 'inst_alpha', name: 'alpha' }],
  totalInstalls: 1,
  ...fields,
})

const renderCard = (value: IDeploymentPlanStage) =>
  render(
    <MemoryRouter>
      <InstallGroupCard
        stage={value}
        groupHref="?panel=group%3Agrp_core"
        onInstallSelect={() => {}}
      />
    </MemoryRouter>
  )

describe('InstallGroupCard', () => {
  test('renders all three membership rules', () => {
    const view = renderCard(stage())
    expect(screen.getByText('A fixed list of 1 install')).toBeTruthy()

    view.rerender(
      <MemoryRouter>
        <InstallGroupCard
          stage={stage({ membership: 'all_installs' })}
          onInstallSelect={() => {}}
        />
      </MemoryRouter>
    )
    expect(screen.getByText('Every install on this branch')).toBeTruthy()

    view.rerender(
      <MemoryRouter>
        <InstallGroupCard
          stage={stage({
            membership: 'label_selector',
            selector: { match_labels: { env: 'prod' } },
          })}
          onInstallSelect={() => {}}
        />
      </MemoryRouter>
    )
    expect(screen.getByText('Installs matching these labels')).toBeTruthy()
  })

  test('renders positive, negative, and wildcard selector chips', () => {
    renderCard(
      stage({
        membership: 'label_selector',
        selector: {
          match_labels: { env: '*', tier: 'a' },
          not_match_labels: { region: 'us-east-1' },
        },
      })
    )

    expect(screen.getByText('any')).toBeTruthy()
    expect(screen.queryByText('*')).toBeNull()
    expect(screen.getByText('tier')).toBeTruthy()
    expect(screen.getByLabelText('not region=us-east-1')).toBeTruthy()
  })

  test('explains dynamic membership only for selector groups', () => {
    const view = renderCard(
      stage({
        membership: 'label_selector',
        selector: { match_labels: { env: 'prod' } },
      })
    )

    expect(screen.getByText(/Membership is resolved at rollout time/)).toBeTruthy()

    view.rerender(
      <MemoryRouter>
        <InstallGroupCard stage={stage()} onInstallSelect={() => {}} />
      </MemoryRouter>
    )
    expect(
      screen.queryByText(/Membership is resolved at rollout time/)
    ).toBeNull()
  })

  test('bounds inline installs and links to the full count', () => {
    const installs = Array.from({ length: 340 }, (_, index) => ({
      id: `inst_${index}`,
      name: `install-${index + 1}`,
    }))
    renderCard(stage({ installs, totalInstalls: 340 }))

    expect(screen.getAllByRole('button')).toHaveLength(MAX_INLINE_INSTALLS)
    expect(
      screen.getByRole('link', { name: 'View all 340 installs' })
    ).toBeTruthy()
    expect(screen.queryByText('install-6')).toBeNull()
  })

  test('renders rollout counts supplied by the resolved stage', () => {
    renderCard(
      stage({
        installs: [{ id: 'inst_alpha' }, { id: 'inst_bravo' }],
        totalInstalls: 12,
        completedInstalls: 10,
        failedInstalls: 2,
        status: 'error',
      })
    )

    expect(screen.getByText('12 installs')).toBeTruthy()
    expect(screen.getByText('10 done')).toBeTruthy()
    expect(screen.getByText('2 failed')).toBeTruthy()
  })

  test('renders approval waiting as a warning and no actions', () => {
    renderCard(stage({ status: 'approval-awaiting' }))

    expect(screen.getByText('Waiting for approval')).toBeTruthy()
    expect(screen.queryByRole('button', { name: /approve/i })).toBeNull()
    expect(screen.queryByRole('button', { name: /deny/i })).toBeNull()
  })
})
