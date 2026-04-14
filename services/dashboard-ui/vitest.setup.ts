import { vi } from 'vitest'
import '@testing-library/jest-dom/vitest'

vi.mock('./client/configs/api', () => ({
  API_URL: process.env.NUON_API_URL || 'http://localhost:8081',
  POLLING_TIMEOUT: 10000,
  POLLING_TIMEOUT_SHORT: 5000,
  POLLING_TIMEOUT_LOGS: 2000,
}))
