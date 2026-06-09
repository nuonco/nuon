export default {
  title: 'Notebooks/CellTerminal',
}

import type { TOTELLog } from '@/types'
import { CellTerminal } from './NotebookCellLogs'

const makeLogs = (bodies: string[], severity = 9): TOTELLog[] =>
  bodies.map((body, i) => ({
    id: `log-${i}`,
    body,
    timestamp: new Date(Date.now() - (bodies.length - i) * 1000).toISOString(),
    severity_number: severity,
    severity_text: severity <= 8 ? 'DEBUG' : severity <= 12 ? 'INFO' : severity <= 16 ? 'WARN' : 'ERROR',
    scope_name: 'oteljob',
    service_name: 'runner',
    log_attributes: { 'nuon.command_output': 'true' },
  }))

const sampleOutput = [
  'NAME                                    READY   STATUS    RESTARTS   AGE',
  'api-server-6d4f8b5c9-x2k4m             1/1     Running   0          3d',
  'worker-7f9a3c1d8-p5n2j                  1/1     Running   0          3d',
  'redis-master-0                          1/1     Running   0          5d',
]

export const Completed = () => (
  <CellTerminal
    logs={makeLogs(sampleOutput)}
    command="kubectl get pods"
    runCreatedAt={new Date(Date.now() - 5000).toISOString()}
    runUpdatedAt={new Date(Date.now() - 2000).toISOString()}
    isRunComplete
    runFailed={false}
  />
)

export const Failed = () => (
  <CellTerminal
    logs={makeLogs(
      ['Error: unable to connect to cluster', 'exit status 1'],
      17
    )}
    command="kubectl get pods"
    runCreatedAt={new Date(Date.now() - 3000).toISOString()}
    runUpdatedAt={new Date(Date.now() - 1000).toISOString()}
    isRunComplete
    runFailed
  />
)

export const Running = () => (
  <CellTerminal
    logs={makeLogs(sampleOutput.slice(0, 2))}
    command="kubectl get pods --watch"
    runCreatedAt={new Date(Date.now() - 10000).toISOString()}
    isRunComplete={false}
    runFailed={false}
  />
)

export const Empty = () => (
  <CellTerminal
    logs={[]}
    command="echo hello"
    runCreatedAt={new Date().toISOString()}
    isRunComplete={false}
    runFailed={false}
    connectionState="connected"
  />
)

export const MixedSeverities = () => (
  <CellTerminal
    logs={[
      ...makeLogs(['Checking dependencies...'], 9),
      ...makeLogs(['Warning: package outdated'], 13).map((l) => ({
        ...l,
        id: 'warn-1',
      })),
      ...makeLogs(['Error: build step failed'], 17).map((l) => ({
        ...l,
        id: 'err-1',
      })),
      ...makeLogs(['[debug] Retrying...'], 5).map((l) => ({
        ...l,
        id: 'dbg-1',
      })),
    ]}
    command="./build.sh"
    runCreatedAt={new Date(Date.now() - 8000).toISOString()}
    runUpdatedAt={new Date(Date.now() - 1000).toISOString()}
    isRunComplete
    runFailed
  />
)
