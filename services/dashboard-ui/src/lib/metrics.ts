import { StatsD, ClientOptions as StatsDOptions } from 'hot-shots'
import { matchRoutePattern } from './route-matcher'

interface LoggingStatsDOptions extends StatsDOptions {
  enabled?: boolean
  loggingEnabled?: boolean
}

class LoggingStatsD {
  public client: StatsD
  public enabled: boolean
  public loggingEnabled: boolean

  constructor(options: LoggingStatsDOptions) {
    const { enabled = true, loggingEnabled = false, ...statsdOptions } = options

    this.client = new StatsD(statsdOptions)
    this.enabled = enabled
    this.loggingEnabled = loggingEnabled
  }

  timing(stat: string, value: number, tags?: any) {
    if (this.loggingEnabled) {
      // eslint-disable-next-line no-console
      console.log('[metric] timing:', stat, value, tags)
    }
    if (this.enabled) {
      this.client.timing(stat, value, tags)
    }
  }

  increment(stat: string, value?: number, tags?: any) {
    if (this.loggingEnabled) {
      // eslint-disable-next-line no-console
      console.log('[metric] increment:', stat, value, tags)
    }

    if (this.enabled) {
      this.client.increment(stat, value, tags)
    }
  }

  gauge(stat: string, value: number, tags?: any) {
    if (this.loggingEnabled) {
      // eslint-disable-next-line no-console
      console.log('[metric] gauge:', stat, value, tags)
    }

    if (this.enabled) {
      this.client.gauge(stat, value, tags)
    }
  }

  histogram(stat: string, value: number, tags?: any) {
    if (this.loggingEnabled) {
      // eslint-disable-next-line no-console
      console.log('[metric] histogram:', stat, value, tags)
    }

    if (this.enabled) {
      this.client.histogram(stat, value, tags)
    }
  }

  close(callback?: () => void) {
    this.client.close(callback)
  }
}

// Create a singleton StatsD client
let statsd: LoggingStatsD

// initialize statsd client and instrument all api calls
export async function setupMetrics() {
  const enabled = !(process.env.DISABLE_METRICS === 'true')
  const loggingEnabled = process.env.LOG_METRICS === 'true'
  const serviceType = process.env.SERVICE_TYPE || 'unknown'
  const serviceDeployment = process.env.SERVICE_DEPLOYMENT || 'unknown'

  // eslint-disable-next-line no-console
  console.log(
    `metrics setup: enabled=${enabled}, loggingEnabled=${loggingEnabled}`
  )

  statsd = new LoggingStatsD({
    host: process.env.HOST_IP,
    port: 8125,
    mock: !enabled,

    enabled: enabled,
    loggingEnabled: loggingEnabled,
    globalTags: {
      service_type: serviceType,
      service_deployment: serviceDeployment,
    },
    errorHandler: (error) => {
      console.error('error setting up metrics:', error)
    },
  })

  // Patch the native http/https modules
  const http = require('http')
  const originalCreateServer = http.Server.prototype.emit

  http.Server.prototype.emit = function (event: string, ...args: any[]) {
    if (event === 'request') {
      const [req, res] = args

      if (req.url) {
        const start = Date.now()
        const endpoint = matchRoutePattern(req.url)

        // Override res.end to capture metrics when response completes
        const originalEnd = res.end
        res.end = function (...endArgs: any[]) {
          try {
            const end = Date.now()
            const status = res.statusCode < 400 ? 'success' : 'error'
            const statusCodeClass = Math.floor(res.statusCode / 100) + 'xx'
            const tags = {
              method: req.method.toLowerCase(),
              endpoint: endpoint,
              status: status,
              status_code_class: statusCodeClass,
            }

            // console.log(tags)
            statsd.timing('ui.request.latency', end - start, tags)
          } catch (error) {
            console.error('error sending metric ui.request.latency:', error)
          }
          return originalEnd.apply(res, endArgs)
        }
      }
    }

    return originalCreateServer.apply(this, [event, ...args])
  }

  // Graceful shutdown
  const shutdown = () => {
    statsd.close(() => {
      process.exit(0) // Do we need this here? Someone else's responsibility?
    })
  }

  process.on('SIGTERM', shutdown)
  process.on('SIGINT', shutdown)
}

// Export statsd instance for manual tracking if needed
export { statsd }
