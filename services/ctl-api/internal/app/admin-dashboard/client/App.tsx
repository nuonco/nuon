import { createBrowserRouter, RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from '@/providers/config-provider'
import { AppLayout } from '@/components/layout/AppLayout'
import { Home } from '@/views/Home'
import { OrgsList } from '@/views/orgs/OrgsList'
import { OrgDetail } from '@/views/orgs/OrgDetail'
import { AccountsList } from '@/views/accounts/AccountsList'
import { AccountDetail } from '@/views/accounts/AccountDetail'
import { InstallsList } from '@/views/installs/InstallsList'
import { InstallDetail } from '@/views/installs/InstallDetail'
import { RunnersList } from '@/views/runners/RunnersList'
import { AllRunners } from '@/views/runners/AllRunners'
import { RunnerDetail } from '@/views/runners/RunnerDetail'
import { QueuesList } from '@/views/queues/QueuesList'
import { QueueDetail } from '@/views/queues/QueueDetail'
import { QueueSignalDetail as QueueSignalDetailView } from '@/views/queues/QueueSignalDetail'
import { QueueEmitterDetail as QueueEmitterDetailView } from '@/views/queues/QueueEmitterDetail'
import { SignalGraphView } from '@/views/queues/SignalGraphView'
import { WorkflowsList } from '@/views/workflows/WorkflowsList'
import { WorkflowDetail } from '@/views/workflows/WorkflowDetail'
import { LogStreams } from '@/views/log-streams/LogStreams'
import { LogStreamDetail } from '@/views/log-streams/LogStreamDetail'
import { QueueSignals } from '@/views/queue-signals/QueueSignals'
import { InFlightSignals } from '@/views/in-flight-signals/InFlightSignals'
import { SignalCatalog } from '@/views/signal-catalog/SignalCatalog'
import { SignalCatalogDetail } from '@/views/signal-catalog/SignalCatalogDetail'
import { Labels } from '@/views/labels/Labels'
import { SandboxMode } from '@/views/sandbox-mode/SandboxMode'
import { TemporalWorkers } from '@/views/temporal-workers/TemporalWorkers'
import { TemporalWorkerDetail } from '@/views/temporal-workers/TemporalWorkerDetail'
import { TemporalWorkflows } from '@/views/temporal-workflows/TemporalWorkflows'
import { SlowQueries } from '@/views/slow-queries/SlowQueries'
import { QueryCatalog } from '@/views/query-catalog/QueryCatalog'
import { RunnerUptime } from '@/views/runner-uptime/RunnerUptime'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

const router = createBrowserRouter([
  {
    element: <AppLayout />,
    children: [
      { index: true, element: <Home /> },
      { path: 'orgs', element: <OrgsList /> },
      { path: 'orgs/:id', element: <OrgDetail /> },
      { path: 'accounts', element: <AccountsList /> },
      { path: 'accounts/:id', element: <AccountDetail /> },
      { path: 'installs', element: <InstallsList /> },
      { path: 'installs/:id', element: <InstallDetail /> },
      { path: 'runners', element: <RunnersList /> },
      { path: 'runners/all', element: <AllRunners /> },
      { path: 'runners/:id', element: <RunnerDetail /> },
      { path: 'queues', element: <QueuesList /> },
      { path: 'queues/:id', element: <QueueDetail /> },
      { path: 'queues/:id/signals/:signalId', element: <QueueSignalDetailView /> },
      { path: 'queues/:id/signals/:signalId/graph', element: <SignalGraphView /> },
      { path: 'queues/:id/emitters/:emitterId', element: <QueueEmitterDetailView /> },
      { path: 'workflows', element: <WorkflowsList /> },
      { path: 'workflows/:workflowId', element: <WorkflowDetail /> },
      { path: 'log-streams', element: <LogStreams /> },
      { path: 'log-streams/:logStreamId', element: <LogStreamDetail /> },
      { path: 'queue-signals', element: <QueueSignals /> },
      { path: 'in-flight-signals', element: <InFlightSignals /> },
      { path: 'signal-catalog', element: <SignalCatalog /> },
      { path: 'signal-catalog/:signalType', element: <SignalCatalogDetail /> },
      { path: 'labels', element: <Labels /> },
      { path: 'sandbox-mode', element: <SandboxMode /> },
      { path: 'temporal-workers', element: <TemporalWorkers /> },
      { path: 'temporal-workers/:namespace', element: <TemporalWorkerDetail /> },
      { path: 'temporal-workflows', element: <TemporalWorkflows /> },
      { path: 'queries', element: <SlowQueries /> },
      { path: 'query-catalog', element: <QueryCatalog /> },
      { path: 'runner-uptime', element: <RunnerUptime /> },
    ],
  },
])

export const App = () => (
  <ConfigProvider>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </ConfigProvider>
)
