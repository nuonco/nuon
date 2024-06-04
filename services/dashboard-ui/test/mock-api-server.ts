import { setupServer } from 'msw/node'
import { handlers } from './mock-api-handlers'
 
export const server = setupServer(...handlers)
