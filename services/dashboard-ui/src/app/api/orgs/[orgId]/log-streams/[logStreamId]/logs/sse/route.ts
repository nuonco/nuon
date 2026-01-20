import { type NextRequest } from 'next/server'
import { getLogStreamLogs } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'logStreamId'>
) {
  const { logStreamId, orgId } = await params

  const encoder = new TextEncoder()

  const stream = new ReadableStream({
    start(controller) {
      let currentOffset: string | undefined = undefined
      let isActive = true

      const pollLogs = async () => {
        if (!isActive) return

        try {
          const response = await getLogStreamLogs({
            logStreamId,
            orgId,
            offset: currentOffset,
            order: 'asc',
          })

          const nextOffset = response.headers?.['x-nuon-api-next']
          if (nextOffset) {
            currentOffset = nextOffset
          }

          if (response.data && response.data.length > 0) {
            await new Promise<void>((resolve) => {
              let logIndex = 0
              const sendNextLog = () => {
                if (logIndex >= response.data.length) {
                  resolve()
                  return
                }

                const singleLog = [response.data[logIndex]]
                const eventData = `data: ${JSON.stringify(singleLog)}\n\n`
                controller.enqueue(encoder.encode(eventData))

                logIndex++
                setTimeout(sendNextLog, 200)
              }
              sendNextLog()
            })
          }

          if (isActive) {
            setTimeout(pollLogs, 1000)
          }
        } catch (error) {
          const errorEvent = `event: error\ndata: ${JSON.stringify({ error: 'Polling failed' })}\n\n`
          controller.enqueue(encoder.encode(errorEvent))

          if (isActive) {
            setTimeout(pollLogs, 5000)
          }
        }
      }

      pollLogs()

      return () => {
        isActive = false
      }
    },
    cancel() {},
  })

  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache, no-store, must-revalidate',
      'Connection': 'keep-alive',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Headers': 'Cache-Control',
    },
  })
}
