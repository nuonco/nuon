import { Route, Routes } from 'react-router'
import { ActionDetail } from './ActionDetail'
import { ConfigPanel } from './ConfigPanel'
import { AppDetail } from './AppDetail'
import { AppResolver } from './AppResolver'
import { AppSetup } from './AppSetup'
import { BranchActivity } from './BranchActivity'
import { BranchCompare } from './BranchCompare'
import { BranchConfig } from './BranchConfig'
import { BranchDetails } from './BranchDetails'
import { BranchOverview } from './BranchOverview'
import { ResourceTable, makeRows } from './ResourceTable'
import { RolesPage } from './RolesPage'
import { SettingsConnections } from './SettingsConnections'
import { SettingsLayout } from './SettingsLayout'
import { TeamPage } from './TeamPage'
import { ComponentDetailRoute } from './ComponentDetailRoute'
import { AppsPage } from './AppsPage'
import { EntityPage, StatePanel } from './EntityPage'
import { HomePage } from './HomePage'
import { InstallActivity } from './InstallActivity'
import { InstallDetail } from './InstallDetail'
import { InstallDetails } from './InstallDetails'
import { InstallOverview } from './InstallOverview'
import { InstallRunner } from './InstallRunner'
import { InstallsPage } from './InstallsPage'
import { Page } from './Page'
import { PlaceholderGrid } from './PlaceholderGrid'
import { RunPage } from './RunPage'
import {
  RunLogs,
  RunOutputs,
  RunSummary,
  RunTrace,
} from './RunDetail'

export const AppRoutes = () => (
  <Routes>
    <Route
      path="/"
      element={
        <Page actions={['Create app']}>
          <HomePage />
        </Page>
      }
    />
    <Route
      path="/apps"
      element={
        <Page crumbs={[{ label: 'Apps' }]} actions={['Create app']}>
          <AppsPage />
        </Page>
      }
    />
    <Route path="/apps/:appId" element={<AppResolver />} />
    <Route
      path="/apps/:appId/setup"
      element={
        <Page
          crumbs={[
            { label: 'Apps', path: '/apps' },
            { label: 'Set up' },
          ]}
        >
          <AppSetup />
        </Page>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId"
      element={
        <AppDetail actions={['Sync']}>
          <BranchOverview />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/config"
      element={
        <AppDetail crumbs={[{ label: 'Config' }]} actions={['Sync config']}>
          <BranchConfig />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/compare/:fromRunId/:toRunId"
      element={
        <Page
          crumbs={[
            { label: 'Apps', path: '/apps' },
            { label: 'Config' },
            { label: 'Compare' },
          ]}
          actions={['Open in GitHub']}
        >
          <BranchCompare />
        </Page>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/activity"
      element={
        <AppDetail crumbs={[{ label: 'Activity' }]}>
          <BranchActivity />
        </AppDetail>
      }
    />

    <Route
      path="/apps/:appId/branches/:branchId/details"
      element={
        <AppDetail crumbs={[{ label: 'Source' }]} actions={['Sync config']}>
          <BranchDetails />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/components/:componentId"
      element={
        <AppDetail crumbs={[{ label: 'Component' }]} actions={['Build']}>
          <EntityPage
            stats={['Status', 'Type', 'Last build', 'Version']}
            historyTitle="Builds"
          >
            <ConfigPanel />
          </EntityPage>
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/stack"
      element={
        <AppDetail crumbs={[{ label: 'Stack' }]}>
          <ConfigPanel title="Stack configuration" />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/sandbox"
      element={
        <AppDetail crumbs={[{ label: 'Sandbox' }]} actions={['Build']}>
          <EntityPage
            stats={['Status', 'Provider', 'Last build', 'Version']}
            historyTitle="Sandbox builds"
          >
            <ConfigPanel />
          </EntityPage>
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/actions/:actionId"
      element={
        <AppDetail crumbs={[{ label: 'Action' }]}>
          <ConfigPanel />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/runbooks/:actionId"
      element={
        <AppDetail crumbs={[{ label: 'Runbook' }]}>
          <ConfigPanel />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/roles/:roleId"
      element={
        <AppDetail crumbs={[{ label: 'Role' }]}>
          <ConfigPanel title="Role configuration" />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/policies/:policyId"
      element={
        <AppDetail crumbs={[{ label: 'Policy' }]}>
          <ConfigPanel title="Policy rule" />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/builds"
      element={
        <AppDetail crumbs={[{ label: 'Builds' }]} actions={['Build all']}>
          <EntityPage
            stats={['Status', 'Components', 'Last build', 'Duration']}
            historyTitle="Builds"
          />
        </AppDetail>
      }
    />
    <Route
      path="/apps/:appId/branches/:branchId/groups/:groupId"
      element={
        <AppDetail crumbs={[{ label: 'Group' }]} actions={['Edit group']}>
          <EntityPage
            stats={['Status', 'Installs', 'Last rollout', 'Order']}
            historyTitle="Rollouts"
          >
            <ResourceTable
              headers={['name & id', 'status', 'location', 'updated']}
              rows={makeRows('ins', 4, 'Install')}
            />
          </EntityPage>
        </AppDetail>
      }
    />

    <Route
      path="/installs"
      element={
        <Page crumbs={[{ label: 'Installs' }]} actions={['Create install']}>
          <InstallsPage />
        </Page>
      }
    />
    <Route
      path="/installs/:installId"
      element={
        <InstallDetail>
          <InstallOverview />
        </InstallDetail>
      }
    />
    <Route
      path="/installs/:installId/activity"
      element={
        <InstallDetail crumbs={[{ label: 'Activity' }]}>
          <InstallActivity />
        </InstallDetail>
      }
    />

    <Route
      path="/installs/:installId/details"
      element={
        <InstallDetail crumbs={[{ label: 'Details' }]} actions={['Edit inputs']}>
          <InstallDetails />
        </InstallDetail>
      }
    />
    <Route
      path="/installs/:installId/stack"
      element={
        <InstallDetail crumbs={[{ label: 'Stack' }]} actions={['Run']}>
          <EntityPage
            stats={['Status', 'Roles', 'Last run', 'Drift']}
            historyTitle="Stack runs"
          >
            <StatePanel />
          </EntityPage>
        </InstallDetail>
      }
    />
    <Route
      path="/installs/:installId/roles"
      element={
        <InstallDetail crumbs={[{ label: 'Roles' }]}>
          <RolesPage />
        </InstallDetail>
      }
    />
    <Route
      path="/installs/:installId/sandbox"
      element={
        <InstallDetail crumbs={[{ label: 'Sandbox' }]} actions={['Reprovision']}>
          <EntityPage
            stats={['Status', 'Version', 'Last run', 'Drift']}
            historyTitle="Provisions"
          >
            <StatePanel />
          </EntityPage>
        </InstallDetail>
      }
    />
    <Route
      path="/installs/:installId/runner"
      element={
        <InstallDetail crumbs={[{ label: 'Runner' }]} actions={['Restart runner']}>
          <InstallRunner />
        </InstallDetail>
      }
    />

    <Route
      path="/installs/:installId/components/:componentId"
      element={<ComponentDetailRoute />}
    />
    <Route
      path="/installs/:installId/components/:componentId/deploys/:runId"
      element={<RunPage kind="deploy" />}
    >
      <Route index element={<RunSummary />} />
      <Route path="logs" element={<RunLogs />} />
      <Route path="trace" element={<RunTrace />} />
      <Route path="outputs" element={<RunOutputs />} />
    </Route>

    <Route
      path="/installs/:installId/actions/:actionId"
      element={
        <InstallDetail crumbs={[{ label: 'Action' }]} actions={['Run action']}>
          <ActionDetail />
        </InstallDetail>
      }
    />
    <Route
      path="/installs/:installId/actions/:actionId/runs/:runId"
      element={<RunPage kind="action" />}
    >
      <Route index element={<RunSummary />} />
      <Route path="logs" element={<RunLogs />} />
      <Route path="trace" element={<RunTrace />} />
      <Route path="outputs" element={<RunOutputs />} />
    </Route>
    <Route
      path="/installs/:installId/runbooks/:actionId"
      element={
        <InstallDetail crumbs={[{ label: 'Runbook' }]} actions={['Run runbook']}>
          <ActionDetail />
        </InstallDetail>
      }
    />
    <Route
      path="/installs/:installId/runbooks/:actionId/runs/:runId"
      element={<RunPage kind="runbook" />}
    >
      <Route index element={<RunSummary />} />
      <Route path="logs" element={<RunLogs />} />
      <Route path="trace" element={<RunTrace />} />
      <Route path="outputs" element={<RunOutputs />} />
    </Route>

    <Route
      path="/team"
      element={
        <Page crumbs={[{ label: 'Team' }]} actions={['Invite']}>
          <TeamPage />
        </Page>
      }
    />
    <Route
      path="/docs"
      element={
        <Page crumbs={[{ label: 'Docs' }]}>
          <PlaceholderGrid rows={3} height="h-[10rem]" />
        </Page>
      }
    />
    <Route path="/settings" element={<SettingsLayout />}>
      <Route index element={<SettingsConnections />} />
      <Route
        path="webhooks"
        element={
          <ResourceTable
            headers={['endpoint', 'events', 'status', 'created']}
            rows={makeRows('whk', 5, 'Webhook')}
            filters={['Events']}
          />
        }
      />
      <Route
        path="triggers"
        element={
          <ResourceTable
            headers={['name & id', 'event', 'status', 'created']}
            rows={makeRows('trg', 4, 'Trigger')}
          />
        }
      />
      <Route
        path="api-tokens"
        element={
          <ResourceTable
            headers={['name & id', 'last used', 'status', 'created']}
            rows={makeRows('tok', 5, 'Token')}
          />
        }
      />
      <Route
        path="service-accounts"
        element={
          <ResourceTable
            headers={['name & id', 'role', 'status', 'created']}
            rows={makeRows('svc', 4, 'Service account')}
          />
        }
      />
      <Route
        path="oidc"
        element={
          <ResourceTable
            headers={['provider', 'subject', 'status', 'created']}
            rows={makeRows('oid', 3, 'Federation')}
          />
        }
      />
    </Route>
  </Routes>
)
