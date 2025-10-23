export async function register() {
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    const { setupMetrics } = await import('./lib/metrics')
    await setupMetrics()
  }
}
