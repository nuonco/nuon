import { type NextRequest } from 'next/server'
import { getLogStreamLogs } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'logStreamId'>
) {
  const { logStreamId, orgId } = await params

  // Set up SSE response headers
  const encoder = new TextEncoder()
  
  const stream = new ReadableStream({
    start(controller) {
      let currentOffset: string | undefined = undefined
      let isActive = true

      // Polling function
      const pollLogs = async () => {
        if (!isActive) return

        try {
          const response = await getLogStreamLogs({
            logStreamId,
            orgId,
            offset: currentOffset,
            order: 'asc',
          })

          if (response.data && response.data.length > 0) {
            // Send each log individually with small delays
            const sendLogWithDelay = (logIndex: number) => {
              if (!isActive || logIndex >= response.data.length) {
                // Update offset after all logs are sent
                const nextOffset = response.headers?.['x-nuon-api-next']
                if (nextOffset) {
                  currentOffset = nextOffset
                }
                return
              }

              // Send single log
              const singleLog = [response.data[logIndex]]
              const eventData = `data: ${JSON.stringify(singleLog)}\n\n`
              controller.enqueue(encoder.encode(eventData))

              // Send next log after delay
              setTimeout(() => sendLogWithDelay(logIndex + 1), 200) // 200ms between logs
            }

            // Start sending logs one by one
            sendLogWithDelay(0)
          } else {
            // No new logs, update offset anyway
            const nextOffset = response.headers?.['x-nuon-api-next']
            if (nextOffset) {
              currentOffset = nextOffset
            }
          }

          // Continue polling if still active
          if (isActive) {
            setTimeout(pollLogs, 1000) // Poll every 2 seconds
          }
        } catch (error) {
          console.error('SSE polling error:', error)
          // Send error event to client
          const errorEvent = `event: error\ndata: ${JSON.stringify({ error: 'Polling failed' })}\n\n`
          controller.enqueue(encoder.encode(errorEvent))
          
          // Retry after longer delay on error
          if (isActive) {
            setTimeout(pollLogs, 5000)
          }
        }
      }

      // Start initial poll
      pollLogs()

      // Cleanup function
      return () => {
        isActive = false
      }
    },
    cancel() {
      // Client disconnected
      // eslint-disable-next-line
      console.log(`SSE connection closed for log stream: ${logStreamId}`)
    }
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
