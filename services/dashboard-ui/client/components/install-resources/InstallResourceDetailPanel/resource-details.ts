import { stringify } from 'yaml'
import type { TInstallResource } from '@/types'

export function parseResourceDetails(details?: string): unknown {
  if (!details) return undefined
  try {
    return JSON.parse(details)
  } catch {
    return undefined
  }
}

const record = (value: unknown): Record<string, unknown> =>
  value !== null && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {}

const records = (value: unknown): Record<string, unknown>[] =>
  Array.isArray(value) ? value.map(record) : []

const scalar = (value: unknown): string =>
  ['string', 'number', 'boolean'].includes(typeof value)
    ? String(value)
    : 'Not reported'

const selector = (value: unknown): string | undefined => {
  const labels = record(value)
  const pairs = Object.entries(labels).map(
    ([key, val]) => `${key}=${scalar(val)}`
  )
  return pairs.length ? pairs.join(', ') : undefined
}

const containerState = (container: Record<string, unknown>): string => {
  const state = record(container.state)
  if (state.waiting) return scalar(record(state.waiting).reason ?? 'Waiting')
  if (state.terminated) {
    const terminated = record(state.terminated)
    return scalar(
      terminated.reason ??
        (terminated.exitCode === 0
          ? 'Completed'
          : terminated.exitCode !== undefined
            ? `ExitCode:${terminated.exitCode}`
            : 'Terminated')
    )
  }
  return state.running ? 'Running' : 'Not reported'
}

// These are display fields from the reported snapshot, not a health assessment.
export function resourceStatusFields(
  resource: TInstallResource
): [string, string][] {
  const details = record(parseResourceDetails(resource.details))
  const spec = record(details.spec)
  const status = record(details.status)
  const fields: [string, unknown][] = []
  if (resource.provider === 'probe' || resource.provider === 'custom') {
    const probe = record(details.probe)
    const result =
      probe.status_code !== undefined
        ? `HTTP ${probe.status_code}`
        : probe.exit_code !== undefined
          ? `Exit code ${probe.exit_code}`
          : resource.message ||
            (resource.health === 'healthy'
              ? resource.kind === 'TCPProbe'
                ? 'Connected'
                : 'Check passed'
              : resource.health === 'unhealthy' ||
                  resource.health === 'degraded'
                ? 'Check failed'
                : undefined)
    fields.push(
      ['Result', result],
      [
        'Latency',
        typeof probe.latency_ms === 'number'
          ? `${probe.latency_ms}ms`
          : undefined,
      ]
    )
  } else if (resource.kind === 'Pod') {
    const containers = records(status.containerStatuses)
    const init = records(status.initContainerStatuses)
    let display = status.reason ?? status.phase
    const initializing = init.findIndex((container) => {
      const state = record(container.state)
      return !(
        record(state.terminated).exitCode === 0 ||
        (state.running && container.ready === true)
      )
    })
    if (initializing >= 0) {
      const state = containerState(init[initializing])
      display =
        state === 'Running' || state === 'Not reported'
          ? `Init:${initializing}/${init.length}`
          : `Init:${state}`
    } else {
      const blocked = containers.find((container) => {
        const state = containerState(container)
        return !['Running', 'Completed', 'Not reported'].includes(state)
      })
      if (blocked) display = containerState(blocked)
      else if (
        containers.length &&
        containers.every(
          (container) => containerState(container) === 'Completed'
        )
      )
        display = 'Completed'
    }
    const all = [...init, ...containers]
    fields.push(
      [
        'Ready',
        containers.length &&
        containers.every((container) => typeof container.ready === 'boolean')
          ? `${containers.filter((container) => container.ready).length} / ${containers.length}`
          : undefined,
      ],
      ['Status', display],
      [
        'Restarts',
        all.length &&
        all.every((container) => typeof container.restartCount === 'number')
          ? all.reduce(
              (sum, container) => sum + Number(container.restartCount),
              0
            )
          : undefined,
      ]
    )
  } else {
    switch (resource.kind) {
      case 'Deployment':
      case 'StatefulSet':
      case 'ReplicaSet':
        fields.push([
          'Ready',
          status.readyReplicas !== undefined
            ? `${status.readyReplicas}${spec.replicas !== undefined ? ` / ${spec.replicas}` : ''}`
            : undefined,
        ])
        break
      case 'DaemonSet':
        fields.push([
          'Ready',
          status.numberReady !== undefined
            ? `${status.numberReady}${status.desiredNumberScheduled !== undefined ? ` / ${status.desiredNumberScheduled}` : ''}`
            : undefined,
        ])
        break
      case 'Service':
        fields.push(
          ['Type', spec.type],
          [
            'Ports',
            records(spec.ports)
              .map(
                (port) =>
                  `${scalar(port.port)} → ${scalar(port.targetPort)}/${scalar(port.protocol)}`
              )
              .join(', ') || undefined,
          ]
        )
        break
      case 'Ingress':
        fields.push(
          [
            'Hosts',
            records(spec.rules)
              .map((rule) => rule.host)
              .filter(Boolean)
              .join(', ') || undefined,
          ],
          ['Class', spec.ingressClassName]
        )
        break
      case 'IngressClass':
        fields.push(['Controller', spec.controller])
        break
      case 'Gateway':
        fields.push(['Class', spec.gatewayClassName])
        break
      case 'GatewayClass':
        fields.push(['Controller', spec.controllerName])
        break
      case 'HorizontalPodAutoscaler':
        fields.push(
          ['Current', status.currentReplicas],
          ['Desired', status.desiredReplicas]
        )
        break
      case 'CronJob':
        fields.push(
          ['Schedule', spec.schedule],
          ['Suspended', spec.suspend],
          [
            'Active',
            Array.isArray(status.active) ? status.active.length : undefined,
          ]
        )
        break
      case 'Job':
        fields.push(
          ['Succeeded', status.succeeded],
          ['Failed', status.failed],
          ['Active', status.active]
        )
        break
      case 'PersistentVolumeClaim':
        fields.push(
          ['Phase', status.phase],
          ['Capacity', record(status.capacity).storage]
        )
        break
      default:
        fields.push(['phase', status.phase])
    }
    records(status.conditions).forEach((condition) => {
      if (typeof condition.type === 'string')
        fields.push([condition.type, condition.status])
    })
    // Route conditions belong to a particular parent; don't flatten their scope.
    records(status.parents).forEach((parent) => {
      const reference = record(parent.parentRef)
      const owner = `${reference.namespace ? `${reference.namespace}/` : ''}${reference.name || 'Parent not reported'}`
      records(parent.conditions).forEach((condition) => {
        if (typeof condition.type === 'string')
          fields.push([`${owner}: ${condition.type}`, condition.status])
      })
    })
    const addresses = records(record(status.loadBalancer).ingress)
      .map((address) => address.hostname ?? address.ip)
      .filter(Boolean)
    if (addresses.length) fields.push(['Address', addresses.join(', ')])
  }
  return fields
    .filter(
      ([, value]) => value !== undefined && value !== null && value !== ''
    )
    .map(([label, value]) => [label, scalar(value)])
}

export type TResourceSection = {
  title: string
  fields?: [string, string][]
  table?: { headers: string[]; rows: string[][] }
  text?: string
}

const fieldLines = (fields: [string, string][]) => {
  const width = Math.max(...fields.map(([label]) => label.length)) + 2
  return fields
    .map(([label, value]) => `${`${label}:`.padEnd(width)} ${value}`)
    .join('\n')
}

export function resourceSections(
  resource: TInstallResource
): TResourceSection[] {
  const parsedDetails = parseResourceDetails(resource.details)
  const details = record(parsedDetails)
  const spec = record(details.spec)
  const status = record(details.status)
  const fields: [string, unknown][] = [['Runner ID', resource.runner_id]]
  const sections: TResourceSection[] = []
  let title = 'Runtime'
  const addTable = (title: string, headers: string[], rows: unknown[][]) =>
    sections.push({
      title,
      table: { headers, rows: rows.map((row) => row.map(scalar)) },
    })
  const addConditions = (
    title: string,
    conditions: Record<string, unknown>[],
    fields?: [string, string][]
  ) => {
    if (!conditions.length) {
      if (fields)
        sections.push({ title, fields, text: 'No conditions reported.' })
      return
    }
    const transitions = conditions.some(
      (condition) => condition.lastTransitionTime
    )
    const generations = conditions.some(
      (condition) => condition.observedGeneration !== undefined
    )
    addTable(
      title,
      [
        'TYPE',
        'STATUS',
        'REASON',
        ...(transitions ? ['LAST TRANSITION'] : []),
        ...(generations ? ['OBSERVED GENERATION'] : []),
      ],
      conditions.map((condition) => [
        condition.type,
        condition.status,
        condition.reason,
        ...(transitions ? [condition.lastTransitionTime] : []),
        ...(generations ? [condition.observedGeneration] : []),
      ])
    )
    const section = sections.at(-1)!
    section.fields = fields
    section.text = conditions
      .filter((condition) => condition.message)
      .map(
        (condition) => `${scalar(condition.type)}: ${scalar(condition.message)}`
      )
      .join('\n\n')
  }
  const addData = (title: string, value: unknown) => {
    const entries = Object.entries(record(value))
    sections.push(
      entries.length &&
        entries.every(([, value]) =>
          ['string', 'number', 'boolean'].includes(typeof value)
        )
        ? {
            title,
            fields: entries.map(([label, value]) => [label, scalar(value)]),
          }
        : { title, text: stringify(value).trim() }
    )
  }

  switch (
    resource.provider === 'kubernetes' &&
    ['HTTPProbe', 'TCPProbe', 'ExecProbe'].includes(resource.kind || '')
      ? undefined
      : resource.kind
  ) {
    case 'Deployment':
    case 'ReplicaSet':
    case 'StatefulSet':
      title = 'Workload'
      fields.push(
        ['Desired replicas', spec.replicas],
        ['Current replicas', status.replicas],
        ['Ready replicas', status.readyReplicas],
        ['Available replicas', status.availableReplicas],
        ['Updated replicas', status.updatedReplicas],
        ['Strategy', record(spec.strategy).type],
        ['Selector', selector(record(spec.selector).matchLabels)]
      )
      break
    case 'DaemonSet':
      title = 'Workload'
      fields.push(
        ['Desired nodes', status.desiredNumberScheduled],
        ['Scheduled nodes', status.currentNumberScheduled],
        ['Ready nodes', status.numberReady],
        ['Available nodes', status.numberAvailable]
      )
      break
    case 'Pod': {
      fields.push(
        ['Node', spec.nodeName],
        ['Phase', status.phase],
        ['Pod IP', status.podIP]
      )
      for (const [label, key] of [
        ['Init containers', 'initContainerStatuses'],
        ['Containers', 'containerStatuses'],
      ]) {
        const containers = records(status[key])
        if (containers.length)
          addTable(
            label,
            ['NAME', 'READY', 'RESTARTS', 'STATE'],
            containers.map((container) => {
              return [
                container.name,
                container.ready,
                container.restartCount,
                containerState(container),
              ]
            })
          )
      }
      break
    }
    case 'Service': {
      title = 'Service'
      fields.push(
        ['Type', spec.type],
        ['Cluster IP', spec.clusterIP],
        ['External name', spec.externalName],
        ['Selector', selector(spec.selector)]
      )
      const ports = records(spec.ports)
      if (ports.length)
        addTable(
          'Ports',
          ['NAME', 'PORT', 'TARGET', 'PROTOCOL'],
          ports.map((port) => [
            port.name,
            port.port,
            port.targetPort,
            port.protocol,
          ])
        )
      break
    }
    case 'Ingress': {
      title = 'Ingress'
      fields.push(['Ingress class', spec.ingressClassName])
      const routes = records(spec.rules).flatMap((rule) =>
        records(record(rule.http).paths).map((path) => {
          const service = record(record(path.backend).service)
          const port = record(service.port)
          return [
            rule.host ?? '*',
            path.path,
            service.name,
            port.name ?? port.number,
          ]
        })
      )
      const service = record(record(spec.defaultBackend).service)
      if (service.name)
        routes.push([
          '*',
          '(default)',
          service.name,
          record(service.port).name ?? record(service.port).number,
        ])
      if (routes.length)
        addTable('Routes', ['HOST', 'PATH', 'SERVICE', 'PORT'], routes)
      break
    }
    case 'Job':
      title = 'Job'
      fields.push(
        ['Completions', spec.completions],
        ['Active', status.active],
        ['Succeeded', status.succeeded],
        ['Failed', status.failed]
      )
      break
    case 'CronJob':
      title = 'Scheduling'
      fields.push(
        ['Schedule', spec.schedule],
        ['Suspended', spec.suspend],
        ['Concurrency policy', spec.concurrencyPolicy],
        ['Last scheduled', status.lastScheduleTime]
      )
      break
    case 'HorizontalPodAutoscaler': {
      title = 'Scaling'
      const target = record(spec.scaleTargetRef)
      fields.push(
        [
          'Scale target',
          target.kind && target.name
            ? `${target.kind}/${target.name}`
            : undefined,
        ],
        ['Minimum replicas', spec.minReplicas],
        ['Maximum replicas', spec.maxReplicas],
        ['Current replicas', status.currentReplicas],
        ['Desired replicas', status.desiredReplicas]
      )
      if (status.currentMetrics)
        addData('Current metrics', status.currentMetrics)
      break
    }
    case 'Gateway': {
      title = 'Gateway'
      fields.push(['Gateway class', spec.gatewayClassName])
      const listeners = records(spec.listeners)
      if (listeners.length)
        addTable(
          'Listeners',
          ['NAME', 'PROTOCOL', 'PORT', 'HOSTNAME'],
          listeners.map((listener) => [
            listener.name,
            listener.protocol,
            listener.port,
            listener.hostname ?? '*',
          ])
        )
      break
    }
    case 'HTTPRoute': {
      title = 'Routing'
      if (Array.isArray(spec.hostnames))
        fields.push(['Hostnames', spec.hostnames.join(', ')])
      const parents = records(spec.parentRefs)
      if (parents.length)
        addTable(
          'Parents',
          ['KIND', 'NAME', 'NAMESPACE'],
          parents.map((parent) => [
            parent.kind ?? 'Gateway',
            parent.name,
            parent.namespace ?? resource.namespace,
          ])
        )
      const backends = records(spec.rules).flatMap((rule) =>
        records(rule.backendRefs)
      )
      if (backends.length)
        addTable(
          'Backends',
          ['KIND', 'NAME', 'NAMESPACE', 'PORT', 'WEIGHT'],
          backends.map((backend) => [
            backend.kind ?? 'Service',
            backend.name,
            backend.namespace ?? resource.namespace,
            backend.port,
            backend.weight ?? 1,
          ])
        )
      const parentStatuses = records(status.parents)
      parentStatuses.forEach((parent, index) => {
        const reference = record(parent.parentRef)
        const identity: [string, unknown][] = [
          ['Kind', reference.kind ?? 'Gateway'],
          ['Name', reference.name],
          ['Namespace', reference.namespace ?? resource.namespace],
          ['API group', reference.group ?? 'gateway.networking.k8s.io'],
          ['Section', reference.sectionName],
          ['Port', reference.port],
          ['Controller', parent.controllerName],
        ]
        addConditions(
          parentStatuses.length > 1
            ? `Parent status · ${index + 1}`
            : 'Parent status',
          records(parent.conditions),
          identity
            .filter(([, value]) => value !== undefined && value !== null)
            .map(([label, value]) => [label, scalar(value)])
        )
      })
      break
    }
    case 'HTTPProbe':
    case 'TCPProbe':
    case 'ExecProbe': {
      title = 'Probe'
      const probe = record(details.probe)
      fields.push(
        ['Target', probe.target],
        [
          'Command',
          Array.isArray(probe.command)
            ? JSON.stringify(probe.command)
            : undefined,
        ],
        ['Exit code', probe.exit_code],
        ['HTTP status', probe.status_code],
        ['Latency (ms)', probe.latency_ms]
      )
      break
    }
    case 'PersistentVolumeClaim':
      title = 'Storage'
      fields.push(
        ['Phase', status.phase],
        ['Volume', spec.volumeName],
        ['Storage class', spec.storageClassName],
        ['Capacity', record(status.capacity).storage]
      )
      break
    default: {
      if (details.spec) addData('Spec', details.spec)
      if (details.status) {
        const { conditions, ...rest } = status
        if (Object.keys(rest).length) addData('Status', rest)
        else if (!conditions) addData('Status', details.status)
      }
    }
  }

  const addresses = records(record(status.loadBalancer).ingress)
    .map((address) => address.hostname ?? address.ip)
    .filter(Boolean)
  if (addresses.length) fields.push(['Load balancer', addresses.join(', ')])
  addConditions('Conditions', records(status.conditions))
  if (details.diagnosis) addData('Diagnosis', details.diagnosis)
  if (resource.details && parsedDetails === undefined)
    sections.push({ title: 'Details', text: resource.details })
  if (
    parsedDetails !== undefined &&
    !details.spec &&
    !details.status &&
    !details.probe &&
    !details.diagnosis
  )
    addData('Details', parsedDetails)
  const present = fields.filter(
    ([, value]) => value !== undefined && value !== null && value !== ''
  )
  if (present.length)
    sections.unshift({
      title,
      fields: present.map(([label, value]) => [label, scalar(value)]),
    })
  if (resource.message)
    sections.push({ title: 'Result', text: resource.message })
  return sections
}

export function describeResource(resource: TInstallResource): string {
  const identity: [string, unknown][] = [
    ['Name', resource.name],
    ['Namespace', resource.namespace],
    ['Kind', resource.kind],
    ['Provider', resource.provider],
    ['API group', resource.api_group],
    ['Last observed', resource.observed_at],
  ]
  return [
    fieldLines(
      identity
        .filter(
          ([, value]) => value !== undefined && value !== null && value !== ''
        )
        .map(([label, value]) => [label, scalar(value)])
    ),
    ...resourceSections(resource).map((section) => {
      const parts = []
      if (section.fields) parts.push(fieldLines(section.fields))
      if (section.table) {
        const values = [section.table.headers, ...section.table.rows]
        const widths = section.table.headers.map((_, index) =>
          Math.max(...values.map((row) => row[index].length))
        )
        parts.push(
          values
            .map((row) =>
              row
                .map((cell, index) => cell.padEnd(widths[index]))
                .join('  ')
                .trimEnd()
            )
            .join('\n')
        )
      }
      if (section.text) parts.push(section.text)
      return `${section.title}:\n${parts.join('\n\n')}`
    }),
  ].join('\n\n')
}
