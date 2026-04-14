import { vi, beforeAll, afterEach, afterAll } from 'vitest'
import '@testing-library/jest-dom/vitest'
import { server } from './test/mock-api-server'

vi.mock('./client/configs/api', () => ({
  API_URL: process.env.NUON_API_URL || 'http://localhost:8081',
  POLLING_TIMEOUT: 10000,
  POLLING_TIMEOUT_SHORT: 5000,
  POLLING_TIMEOUT_LOGS: 2000,
}))
 
Object.defineProperty(window, 'location', {
  value: { ...window.location, reload: vi.fn() },
})

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
