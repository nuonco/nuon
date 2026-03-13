interface ITopologyConnectorProps {
  variant?: 'straight' | 'branch'
  count?: number
}

const NODE_W = 200
const NODE_GAP = 12
const STEP = NODE_W + NODE_GAP // 212px per slot

// Arrowhead tip overlaps the card border by this amount (negative margin)
const OVERLAP = 4

export function TopologyConnector({
  variant = 'straight',
  count = 1,
}: ITopologyConnectorProps) {

  // ── straight / single-node ────────────────────────────────────────────────
  if (variant === 'straight' || count <= 1) {
    const w = 20
    const h = 28
    const cx = w / 2
    const arrowTip = h + OVERLAP

    return (
      <svg
        width={w}
        height={h}
        viewBox={`0 0 ${w} ${arrowTip}`}
        fill="none"
        className="text-cool-grey-400 dark:text-dark-grey-500 overflow-visible block"
        style={{ marginBottom: -OVERLAP }}
      >
        {/* exit dot at top */}
        <circle cx={cx} cy={1} r="2" fill="currentColor" />
        {/* stem */}
        <line x1={cx} y1={1} x2={cx} y2={arrowTip - 8} stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
        {/* arrowhead */}
        <polygon
          points={`${cx - 4},${arrowTip - 8} ${cx + 4},${arrowTip - 8} ${cx},${arrowTip}`}
          fill="currentColor"
        />
      </svg>
    )
  }

  // ── branch (fan out to multiple nodes) ───────────────────────────────────
  const svgW = count * STEP - NODE_GAP
  const junctionY = 12
  const dropH = 24          // height of each vertical drop below junction
  const arrowH = 8          // arrowhead height
  const svgH = junctionY + dropH
  const arrowTip = svgH + OVERLAP

  return (
    <svg
      width={svgW}
      height={svgH}
      viewBox={`0 0 ${svgW} ${arrowTip}`}
      fill="none"
      className="text-cool-grey-400 dark:text-dark-grey-500 overflow-visible block"
      style={{ minWidth: svgW, marginBottom: -OVERLAP }}
    >
      {/* vertical entry from parent node */}
      <line
        x1={svgW / 2} y1={0}
        x2={svgW / 2} y2={junctionY}
        stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"
      />

      {/* junction dot */}
      <circle cx={svgW / 2} cy={junctionY} r="2.5" fill="currentColor" />

      {/* horizontal bus bar */}
      <line
        x1={NODE_W / 2} y1={junctionY}
        x2={svgW - NODE_W / 2} y2={junctionY}
        stroke="currentColor" strokeWidth="1.5"
      />

      {/* drop + arrowhead per node */}
      {Array.from({ length: count }).map((_, i) => {
        const x = NODE_W / 2 + i * STEP
        return (
          <g key={i}>
            {/* small dot on the bar at each branch point */}
            <circle cx={x} cy={junctionY} r="2" fill="currentColor" />
            {/* stem drop */}
            <line
              x1={x} y1={junctionY}
              x2={x} y2={arrowTip - arrowH}
              stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"
            />
            {/* arrowhead */}
            <polygon
              points={`${x - 4},${arrowTip - arrowH} ${x + 4},${arrowTip - arrowH} ${x},${arrowTip}`}
              fill="currentColor"
            />
          </g>
        )
      })}
    </svg>
  )
}
