import { createMiddleware } from '@mswjs/http-middleware'
import cors from 'cors'
import express from 'express'
import { handlers } from './mock-api-handlers'
import { nextProxyHandlers } from './mock-api-server'

const mockServer = express()
mockServer.use(cors())
mockServer.use(express.json())
mockServer.use(createMiddleware(...handlers, ...nextProxyHandlers))

mockServer.use((req, res) => {
  const errorMessage = `Mock for ${req.url} not found`

  // eslint-disable-next-line
  console.error(errorMessage)
  res.status(404).send({ error: errorMessage })
})

mockServer.listen(3030, () => {
  // eslint-disable-next-line
  console.log('mock api server running at localhost:3030')
})
