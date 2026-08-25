import { InstallOperationRules } from './InstallOperationRules'
import { makeOperationRules } from './schema'

export default {
  title: 'Operation queues/Install operation rules',
}

const noop = () => {}

export const Default = () => (
  <InstallOperationRules
    installName="production-us-west"
    timezone="America/Los_Angeles"
    rules={makeOperationRules({
      deployments: {
        enabled: true,
        cadence: 'weekly',
        daysOfWeek: ['tue', 'thu'],
        startTime: '10:00',
        endTime: '16:00',
        outsideWindowPolicy: 'queue',
      },
      'break-glass': {
        enabled: true,
        cadence: 'weekly',
        daysOfWeek: ['mon', 'tue', 'wed', 'thu', 'fri'],
        startTime: '08:00',
        endTime: '18:00',
        outsideWindowPolicy: 'approval',
      },
    })}
    onCancel={noop}
    onSave={noop}
  />
)

export const AllDisabled = () => (
  <InstallOperationRules
    installName="production-us-west"
    timezone="UTC"
    rules={makeOperationRules()}
    onCancel={noop}
    onSave={noop}
  />
)

export const AllEnabled = () => (
  <InstallOperationRules
    installName="production-us-west"
    timezone="UTC"
    rules={makeOperationRules({
      actions: {
        enabled: true,
        cadence: 'weekly',
        daysOfWeek: ['mon', 'wed', 'fri'],
        outsideWindowPolicy: 'queue',
      },
      runbooks: {
        enabled: true,
        cadence: 'monthly',
        dayOfMonth: '15',
        startTime: '01:00',
        endTime: '05:00',
        outsideWindowPolicy: 'approval',
      },
      'sandbox-updates': {
        enabled: true,
        cadence: 'monthly',
        dayOfMonth: 'last',
        startTime: '22:00',
        endTime: '23:30',
        outsideWindowPolicy: 'reject',
      },
      deployments: {
        enabled: true,
        cadence: 'weekly',
        daysOfWeek: ['tue', 'thu'],
        outsideWindowPolicy: 'queue',
      },
      'break-glass': {
        enabled: true,
        cadence: 'anytime',
      },
    })}
    onCancel={noop}
    onSave={noop}
  />
)

export const WeeklyChangeWindow = () => (
  <InstallOperationRules
    installName="production-eu-central"
    timezone="Europe/Berlin"
    rules={makeOperationRules({
      deployments: {
        enabled: true,
        cadence: 'weekly',
        daysOfWeek: ['wed'],
        startTime: '02:00',
        endTime: '06:00',
        outsideWindowPolicy: 'queue',
      },
      'sandbox-updates': {
        enabled: true,
        cadence: 'weekly',
        daysOfWeek: ['wed'],
        startTime: '02:00',
        endTime: '06:00',
        outsideWindowPolicy: 'reject',
      },
    })}
    onCancel={noop}
    onSave={noop}
  />
)

export const BreakGlassApproval = () => (
  <InstallOperationRules
    installName="production-us-east"
    timezone="America/New_York"
    rules={makeOperationRules({
      'break-glass': {
        enabled: true,
        cadence: 'monthly',
        dayOfMonth: '1',
        startTime: '00:00',
        endTime: '23:59',
        outsideWindowPolicy: 'approval',
      },
    })}
    onCancel={noop}
    onSave={noop}
  />
)

export const IncompleteWindow = () => (
  <InstallOperationRules
    installName="production-us-west"
    timezone="UTC"
    rules={makeOperationRules({
      actions: {
        enabled: true,
        cadence: 'weekly',
        daysOfWeek: [],
        outsideWindowPolicy: 'reject',
      },
    })}
    onCancel={noop}
    onSave={noop}
  />
)
