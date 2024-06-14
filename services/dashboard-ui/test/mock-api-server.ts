import { delay, http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import {
  handlers,
  getGetOrg200Response,
  getGetInstall200Response,
  getGetInstallComponents200Response,
} from './mock-api-handlers'

const nextProxyHandlers = [
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
]

export const server = setupServer(...handlers, ...nextProxyHandlers)
