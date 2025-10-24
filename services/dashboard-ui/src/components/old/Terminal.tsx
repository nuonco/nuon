/* 'use client'
 *
 * import React, { useEffect, useRef, type FC } from 'react'
 * import { Terminal as XTerm } from '@xterm/xterm'
 * import '@xterm/xterm/css/xterm.css'
 *
 * // TODO(nnnnat): not ready for prime time
 * export const ExperimentalTerminal: FC = () => {
 *   const terminalRef = useRef(null)
 *
 *   useEffect(() => {
 *     const terminal = new XTerm()
 *     terminal.open(terminalRef.current)
 *
 *     return () => {
 *       terminal.dispose()
 *     }
 *   }, [])
 *
 *   return (
 *     <div className="rounded overflow-hidden">
 *       <div ref={terminalRef} />
 *     </div>
 *   )
 * } */
