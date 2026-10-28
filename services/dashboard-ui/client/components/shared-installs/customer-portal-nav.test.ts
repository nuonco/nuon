import { describe, expect, it } from 'bun:test'
import { customerPortalNav } from './customer-portal-nav'

describe('customerPortalNav', () => {
  it('always includes overview and deployments', () => {
    const links = customerPortalNav([])
    expect(links.map((link) => ('text' in link ? link.text : null))).toEqual([
      'Overview',
      'Deployments',
    ])
  })

  it('adds scoped tabs in a stable order', () => {
    const links = customerPortalNav([
      'operations',
      'approvals',
      'resources',
      'readme',
    ])
    expect(links.map((link) => ('text' in link ? link.text : null))).toEqual([
      'Overview',
      'Deployments',
      'Approvals',
      'README',
      'Resources',
      'Operations',
    ])
  })

  it('omits scopes that were not granted', () => {
    const links = customerPortalNav(['readme', 'resources'])
    expect(links.map((link) => ('text' in link ? link.text : null))).toEqual([
      'Overview',
      'Deployments',
      'README',
      'Resources',
    ])
  })
})
