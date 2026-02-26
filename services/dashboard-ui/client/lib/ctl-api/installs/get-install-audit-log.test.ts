import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallAuditLog } from './get-install-audit-log'

describe('getInstallAuditLog should handle response status codes from GET installs/:id/audit_logs endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'
  const start = '2024-01-01T00:00:00Z'
  const end = '2024-01-31T23:59:59Z'

  test('200 status', async () => {
    const { status } = await getInstallAuditLog({
      installId,
      orgId,
      start,
      end,
    })
    expect(status).toBe(200)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallAuditLog({
      installId,
      orgId,
      start,
      end,
    })

    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
    })
  })
})
