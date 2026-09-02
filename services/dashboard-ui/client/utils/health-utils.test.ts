import { describe, expect, test } from 'bun:test'
import {
  bearsHealthVerdict,
  compareHealthSeverityDesc,
  healthSeverity,
  isFailingHealth,
  worstHealth,
} from './health-utils'

describe('health-utils', () => {
  describe('healthSeverity', () => {
    test('ranks assessed statuses worst-first', () => {
      expect(healthSeverity('unhealthy')).toBeGreaterThan(
        healthSeverity('degraded')
      )
      expect(healthSeverity('degraded')).toBeGreaterThan(
        healthSeverity('progressing')
      )
      expect(healthSeverity('progressing')).toBeGreaterThan(
        healthSeverity('healthy')
      )
    })

    test('unknown and not-applicable share zero severity', () => {
      expect(healthSeverity('unknown')).toBe(0)
      expect(healthSeverity('not-applicable')).toBe(0)
      expect(healthSeverity('')).toBe(0)
      expect(healthSeverity(undefined)).toBe(0)
    })

    test('an unrecognized status never outranks an assessed one', () => {
      expect(healthSeverity('wat')).toBeLessThan(healthSeverity('healthy'))
    })
  })

  describe('bearsHealthVerdict', () => {
    test('only assessed statuses bear a verdict', () => {
      expect(bearsHealthVerdict('healthy')).toBe(true)
      expect(bearsHealthVerdict('progressing')).toBe(true)
      expect(bearsHealthVerdict('degraded')).toBe(true)
      expect(bearsHealthVerdict('unhealthy')).toBe(true)
      expect(bearsHealthVerdict('unknown')).toBe(false)
      expect(bearsHealthVerdict('not-applicable')).toBe(false)
      expect(bearsHealthVerdict(undefined)).toBe(false)
    })
  })

  describe('isFailingHealth', () => {
    test('only unhealthy and degraded are failing', () => {
      expect(isFailingHealth('unhealthy')).toBe(true)
      expect(isFailingHealth('degraded')).toBe(true)
      expect(isFailingHealth('progressing')).toBe(false)
      expect(isFailingHealth('healthy')).toBe(false)
      expect(isFailingHealth('unknown')).toBe(false)
    })
  })

  describe('worstHealth', () => {
    test('returns the worst assessed status', () => {
      expect(worstHealth(['healthy', 'degraded', 'progressing'])).toBe(
        'degraded'
      )
      expect(worstHealth(['degraded', 'unhealthy'])).toBe('unhealthy')
    })

    test('an assessed status always beats unknown and not-applicable', () => {
      expect(worstHealth(['unknown', 'healthy', 'not-applicable'])).toBe(
        'healthy'
      )
    })

    test('falls back to unknown when nothing was assessed', () => {
      expect(worstHealth(['unknown', 'not-applicable'])).toBe('unknown')
      expect(worstHealth([])).toBe('unknown')
    })

    test('falls back to not-applicable only when nothing tried', () => {
      expect(worstHealth(['not-applicable', 'not-applicable'])).toBe(
        'not-applicable'
      )
    })
  })

  describe('compareHealthSeverityDesc', () => {
    test('sorts worst-first', () => {
      expect(
        ['healthy', 'unhealthy', 'unknown', 'degraded'].sort(
          compareHealthSeverityDesc
        )
      ).toEqual(['unhealthy', 'degraded', 'healthy', 'unknown'])
    })
  })
})
