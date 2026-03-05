import { describe, expect, test } from 'vitest'
import { getStatusTheme, getStatusIconVariant } from './status-utils'

describe('status-utils', () => {
  describe('getStatusTheme', () => {
    test('should return success theme for success statuses', () => {
      expect(getStatusTheme('active')).toBe('success')
      expect(getStatusTheme('ok')).toBe('success')
      expect(getStatusTheme('finished')).toBe('success')
      expect(getStatusTheme('healthy')).toBe('success')
      expect(getStatusTheme('connected')).toBe('success')
      expect(getStatusTheme('approved')).toBe('success')
      expect(getStatusTheme('success')).toBe('success')
    })

    test('should return error theme for error statuses', () => {
      expect(getStatusTheme('failed')).toBe('error')
      expect(getStatusTheme('error')).toBe('error')
      expect(getStatusTheme('bad')).toBe('error')
      expect(getStatusTheme('access-error')).toBe('error')
      expect(getStatusTheme('access_error')).toBe('error')
      expect(getStatusTheme('timed-out')).toBe('error')
      expect(getStatusTheme('unknown')).toBe('error')
      expect(getStatusTheme('unhealthy')).toBe('error')
      expect(getStatusTheme('not connected')).toBe('error')
      expect(getStatusTheme('not-connected')).toBe('error')
    })

    test('should return warn theme for warning statuses', () => {
      expect(getStatusTheme('approval-denied')).toBe('warn')
      expect(getStatusTheme('approval-awaiting')).toBe('warn')
      expect(getStatusTheme('cancelled')).toBe('warn')
      expect(getStatusTheme('outdated')).toBe('warn')
      expect(getStatusTheme('warn')).toBe('warn')
    })

    test('should return info theme for info statuses', () => {
      expect(getStatusTheme('executing')).toBe('info')
      expect(getStatusTheme('waiting')).toBe('info')
      expect(getStatusTheme('started')).toBe('info')
      expect(getStatusTheme('in-progress')).toBe('info')
      expect(getStatusTheme('building')).toBe('info')
      expect(getStatusTheme('queued')).toBe('info')
      expect(getStatusTheme('planning')).toBe('info')
      expect(getStatusTheme('provisioning')).toBe('info')
      expect(getStatusTheme('syncing')).toBe('info')
      expect(getStatusTheme('deploying')).toBe('info')
      expect(getStatusTheme('available')).toBe('info')
      expect(getStatusTheme('pending-approval')).toBe('info')
      expect(getStatusTheme('info')).toBe('info')
      expect(getStatusTheme('retried')).toBe('info')
    })

    test('should return neutral theme for neutral statuses', () => {
      expect(getStatusTheme('noop')).toBe('neutral')
      expect(getStatusTheme('inactive')).toBe('neutral')
      expect(getStatusTheme('pending')).toBe('neutral')
      expect(getStatusTheme('offline')).toBe('neutral')
      expect(getStatusTheme('Not deployed')).toBe('neutral')
      expect(getStatusTheme('No build')).toBe('neutral')
      expect(getStatusTheme('not-attempted')).toBe('neutral')
      expect(getStatusTheme('deprovisioned')).toBe('neutral')
      expect(getStatusTheme('skeleton')).toBe('neutral')
    })

    test('should return brand theme for brand statuses', () => {
      expect(getStatusTheme('special')).toBe('brand')
      expect(getStatusTheme('brand')).toBe('brand')
    })

    test('should return neutral theme for unknown statuses', () => {
      expect(getStatusTheme('unknown-status')).toBe('neutral')
      expect(getStatusTheme('')).toBe('neutral')
    })
  })

  describe('getStatusIconVariant', () => {
    test('should return CheckCircle for success statuses', () => {
      expect(getStatusIconVariant('active')).toBe('CheckCircle')
      expect(getStatusIconVariant('ok')).toBe('CheckCircle')
      expect(getStatusIconVariant('finished')).toBe('CheckCircle')
      expect(getStatusIconVariant('approved')).toBe('CheckCircle')
    })

    test('should return XCircle for error statuses', () => {
      expect(getStatusIconVariant('failed')).toBe('XCircle')
      expect(getStatusIconVariant('error')).toBe('XCircle')
      expect(getStatusIconVariant('bad')).toBe('XCircle')
      expect(getStatusIconVariant('unhealthy')).toBe('XCircle')
    })

    test('should return Warning for warning statuses', () => {
      expect(getStatusIconVariant('approval-denied')).toBe('Warning')
      expect(getStatusIconVariant('cancelled')).toBe('Warning')
      expect(getStatusIconVariant('outdated')).toBe('Warning')
    })

    test('should return Loading for info statuses', () => {
      expect(getStatusIconVariant('executing')).toBe('Loading')
      expect(getStatusIconVariant('building')).toBe('Loading')
      expect(getStatusIconVariant('deploying')).toBe('Loading')
    })

    test('should return ClockCountdown for neutral statuses', () => {
      expect(getStatusIconVariant('pending')).toBe('ClockCountdown')
      expect(getStatusIconVariant('inactive')).toBe('ClockCountdown')
      expect(getStatusIconVariant('offline')).toBe('ClockCountdown')
    })

    test('should return special icons for specific statuses', () => {
      expect(getStatusIconVariant('user-skipped')).toBe('MinusCircle')
      expect(getStatusIconVariant('retried')).toBe('Repeat')
      expect(getStatusIconVariant('special')).toBe('Prohibit')
      expect(getStatusIconVariant('not-attempted')).toBe('Prohibit')
      expect(getStatusIconVariant('skeleton')).toBe('none')
    })

    test('should return ClockCountdown for unknown statuses', () => {
      expect(getStatusIconVariant('unknown-status')).toBe('ClockCountdown')
      expect(getStatusIconVariant('')).toBe('ClockCountdown')
    })
  })
})
