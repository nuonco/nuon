import { delay, http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import {
  handlers,
  getGetBuild200Response,
  getGetOrg200Response,
  getGetInstall200Response,
  getGetInstallComponents200Response,
  getGetInstallDeploy200Response,
  getGetInstallSandboxRuns200Response,
} from './mock-api-handlers'

export const nextProxyHandlers = [
  http.get('/api/:orgId', async () => {
    await delay(300)
    return HttpResponse.json(getGetOrg200Response(), {
      status: 200,
    })
  }),

  http.get('/api/:orgId/installs/:installId', async () => {
    await delay(300)
    return HttpResponse.json(getGetInstall200Response(), {
      status: 200,
    })
  }),

  http.get(
    '/api/:orgId/installs/:installId/components/:installComponentId',
    async () => {
      await delay(300)
      return HttpResponse.json(getGetInstallComponents200Response()[0], {
        status: 200,
      })
    }
  ),

  http.get('/api/:orgId/components/:componentId/builds/:buildId', async () => {
    await delay(300)
    return HttpResponse.json(getGetBuild200Response(), {
      status: 200,
    })
  }),

  http.get('/api/:orgId/installs/:installId/deploys/:deployId', async () => {
    await delay(300)
    return HttpResponse.json(getGetInstallDeploy200Response(), {
      status: 200,
    })
  }),

  http.get('/api/:orgId/installs/:installId/runs/:runId', async () => {
    await delay(300)
    return HttpResponse.json(getGetInstallSandboxRuns200Response()[0], {
      status: 200,
    })
  }),
]

export const server = setupServer(...handlers)
