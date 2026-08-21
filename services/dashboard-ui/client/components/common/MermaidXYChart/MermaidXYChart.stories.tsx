import { MermaidXYChart } from './MermaidXYChart'

export default {
  title: 'Common/MermaidXYChart',
}

const docsSample = `xychart-beta
    title "Product Sales (2026)"
    x-axis [Jan, Feb, Mar, Apr, May, Jun]
    y-axis "Revenue (in USD)" 0 --> 5000
    bar [1500, 2500, 3200, 2800, 4200, 4800]
    line [1500, 2500, 3200, 2800, 4200, 4800]`

export const BarAndLine = () => <MermaidXYChart code={docsSample} />

const runbookMetric = `xychart-beta
    title "CPU %"
    line [12.62, 14.19, 12.75, 12.73, 12.62, 12.82, 12.59, 13.66, 13.78, 13.13, 12.59, 12.64, 12.73, 12.95, 12.59, 13.5, 12.63, 12.82]`

export const RunbookMetric = () => <MermaidXYChart code={runbookMetric} />

const unparseable = `xychart-beta
    pie [1, 2, 3]`

export const FallbackToSource = () => <MermaidXYChart code={unparseable} />
