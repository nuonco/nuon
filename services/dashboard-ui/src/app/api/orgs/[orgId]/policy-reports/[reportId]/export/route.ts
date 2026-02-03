import { type NextRequest } from 'next/server'
import { API_URL } from '@/configs/api'
import { getAccessToken } from '@/lib/auth-server'
import { type TPolicyReportExportFormat } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  req: NextRequest,
  { params }: TRouteProps<'orgId' | 'reportId'>
) {
  const { reportId, orgId } = await params
  const format =
    (req.nextUrl.searchParams.get('format') as TPolicyReportExportFormat) ||
    'opa'

  try {
    const accessToken = await getAccessToken()
    const response = await fetch(
      `${API_URL}/v1/policy-reports/${reportId}/export?format=${format}`,
      {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'X-Nuon-Org-ID': orgId,
        },
      }
    )

    if (!response.ok) {
      const errorText = await response.text()
      return new Response(errorText, { status: response.status })
    }

    const content = await response.arrayBuffer()
    const contentType =
      response.headers.get('content-type') ||
      (format === 'pdf' ? 'application/pdf' : 'application/json')
    const contentDisposition =
      response.headers.get('content-disposition') ||
      `attachment; filename="policy-report.${format === 'pdf' ? 'pdf' : 'json'}"`

    return new Response(content, {
      status: 200,
      headers: {
        'Content-Type': contentType,
        'Content-Disposition': contentDisposition,
      },
    })
  } catch (error) {
    return new Response(`Export failed: ${error instanceof Error ? error.message : 'unknown error'}`,
      { status: 500 }
    )
  }
}
