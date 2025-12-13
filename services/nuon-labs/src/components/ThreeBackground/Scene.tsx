'use client'

import { useRef, useMemo, Suspense } from 'react'
import { Canvas, useFrame } from '@react-three/fiber'
import * as THREE from 'three'
import { SVGLoader } from 'three/addons/loaders/SVGLoader.js'

// Nuon logo SVG path data
const LOGO_SVG = `
<svg viewBox="0 0 1080 1080">
  <path d="M704.25,72.41L519.57,175.78v167.36l-149.5-83.72h-0.09l-178.2,99.74v548.81l178.12,99.74h0.09l191.24-107.09
    V740.61l143.02,80.01l184.68-103.37V175.78L704.25,72.41z M233.52,382.52l136.38-76.29h0.09l149.5,83.64v280.64L233.52,510.5V382.52
    z M519.48,877.25L369.9,960.89L233.52,884.6V557.23l285.96,160.01V877.25z M847.19,693.79L704.25,773.8l-142.94-79.92V413.24
    l285.96,160.01v120.54H847.19z M847.19,526.52L561.22,366.51V199.15l143.02-80.01l142.94,80.01V526.52z"/>
</svg>
`

const Logo3D = () => {
  const groupRef = useRef<THREE.Group>(null)

  // Parse SVG and create extruded geometry
  const shapes = useMemo(() => {
    const loader = new SVGLoader()
    const svgData = loader.parse(LOGO_SVG)
    return svgData.paths.flatMap((path) => SVGLoader.createShapes(path))
  }, [])

  // Create gradient material
  const material = useMemo(() => {
    return new THREE.MeshStandardMaterial({
      color: '#8040BF',
      metalness: 0.3,
      roughness: 0.4,
      emissive: '#3A00FF',
      emissiveIntensity: 0.2,
    })
  }, [])

  useFrame((state) => {
    if (!groupRef.current) return
    // Slow rotation
    groupRef.current.rotation.y = state.clock.elapsedTime * 0.15
  })

  const extrudeSettings = {
    depth: 80,
    bevelEnabled: true,
    bevelThickness: 8,
    bevelSize: 5,
    bevelOffset: 0,
    bevelSegments: 3,
  }

  return (
    <group ref={groupRef} position={[0, 0, 0]} scale={[0.006, -0.006, 0.006]}>
      {/* Center the logo */}
      <group position={[-540, -540, -15]}>
        {shapes.map((shape, index) => (
          <mesh key={index} material={material}>
            <extrudeGeometry args={[shape, extrudeSettings]} />
          </mesh>
        ))}
      </group>
    </group>
  )
}

export const Scene = () => {
  return (
    <div
      className="fixed inset-0 -z-10"
      style={{ pointerEvents: 'none', width: '100vw', height: '100vh' }}
    >
      {/* Animated purple shimmer background */}
      <div
        className="absolute inset-0 opacity-60"
        style={{
          background: `
            radial-gradient(ellipse at 50% 50%, rgba(128, 64, 191, 0.3) 0%, transparent 50%),
            radial-gradient(ellipse at 30% 30%, rgba(58, 0, 255, 0.15) 0%, transparent 40%),
            radial-gradient(ellipse at 70% 70%, rgba(247, 37, 133, 0.15) 0%, transparent 40%)
          `,
          animation: 'shimmer 8s ease-in-out infinite',
        }}
      />
      <style>{`
        @keyframes shimmer {
          0%, 100% { opacity: 0.4; transform: scale(1); }
          50% { opacity: 0.7; transform: scale(1.05); }
        }
      `}</style>
      <Canvas
        camera={{
          position: [0, 0, 12],
          fov: 45,
          near: 0.1,
          far: 100,
        }}
        style={{
          background: 'transparent',
          width: '100%',
          height: '100%',
        }}
        gl={{
          alpha: true,
          antialias: true,
          powerPreference: 'default',
          failIfMajorPerformanceCaveat: false,
        }}
        onCreated={({ gl }) => {
          gl.setClearColor(0x000000, 0)
        }}
      >
        <Suspense fallback={null}>
          {/* Lighting */}
          <ambientLight intensity={0.4} />
          <directionalLight position={[5, 5, 5]} intensity={1} color="#ffffff" />
          <directionalLight position={[-5, -5, -5]} intensity={0.3} color="#F72585" />
          <pointLight position={[0, 0, 10]} intensity={0.5} color="#4CC9F0" />

          {/* Logo */}
          <Logo3D />
        </Suspense>
      </Canvas>
    </div>
  )
}

export default Scene
