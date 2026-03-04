import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallAuditLog } from './get-install-audit-log'

describe('getInstallAuditLog should handle response status codes from GET installs/:id/audit_logs endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'
  const start = '2024-01-01T00:00:00Z'
  const end = '2024-01-31T23:59:59Z'

  test('200 status', async () => {
    const result = await getInstallAuditLog({ installId, orgId, start, end })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getInstallAuditLog({ installId, orgId, start, end })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
    })
  })
})
