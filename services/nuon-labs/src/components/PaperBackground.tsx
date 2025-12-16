/** @paper-design/shaders-react@0.0.68 */

'use client'

import { StaticRadialGradient } from '@paper-design/shaders-react';
import { useEffect, useState, useRef } from 'react';

/**
 * Code exported from Paper
 * https://app.paper.design/file/01KCHM60V8618F2RX0MR2CNT7K?node=01KCHM84579GQ63DDT5HCR3YD5
 * on Dec 15, 2025 at 2:18 PM.
 */

// Radius range: 0.3 (start) -> 1.3 (60%) -> 3.0 (120%)
const RADIUS_START = 0.3;
const RADIUS_60 = 1.3;  // 60% point (end of intro animation)
const RADIUS_100 = 3.0; // 120% point (full scroll)

export default function PaperBackground() {
  const [radius, setRadius] = useState(RADIUS_START);
  const [introComplete, setIntroComplete] = useState(false);
  const animationRef = useRef<number | null>(null);
  const startTimeRef = useRef<number | null>(null);
  const textTriggeredRef = useRef(false);

  // Scroll to top on mount
  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  // Intro animation: 0 -> 60 on page load
  useEffect(() => {
    const introDuration = 1800; // 1.8 seconds for intro
    const textTriggerPoint = 0.5; // Show text at 50% of animation

    const animateIntro = (timestamp: number) => {
      if (startTimeRef.current === null) {
        startTimeRef.current = timestamp;
      }
      
      const elapsed = timestamp - startTimeRef.current;
      const progress = Math.min(elapsed / introDuration, 1);
      
      // Ease-out cubic for smooth deceleration
      const eased = 1 - Math.pow(1 - progress, 3);
      const newRadius = RADIUS_START + eased * (RADIUS_60 - RADIUS_START);
      setRadius(newRadius);

      // Trigger text fade-in at 50% progress (synced with background)
      if (progress >= textTriggerPoint && !textTriggeredRef.current) {
        textTriggeredRef.current = true;
        window.dispatchEvent(new CustomEvent('intro-complete'));
      }
      
      if (progress < 1) {
        animationRef.current = requestAnimationFrame(animateIntro);
      } else {
        // Intro complete
        setIntroComplete(true);
      }
    };

    // Small delay before starting animation
    const timeout = setTimeout(() => {
      animationRef.current = requestAnimationFrame(animateIntro);
    }, 200);

    return () => {
      clearTimeout(timeout);
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, []);

  // Scroll handler: 60 -> 100 after intro completes
  useEffect(() => {
    if (!introComplete) return;

    const handleScroll = () => {
      // Calculate scroll progress (0 to 1)
      const scrollHeight = document.documentElement.scrollHeight - window.innerHeight;
      const scrollProgress = scrollHeight > 0 ? window.scrollY / scrollHeight : 0;
      
      // Map scroll progress to radius (60% to 100% = RADIUS_60 to RADIUS_100)
      const eased = 1 - Math.pow(1 - scrollProgress, 2); // ease-out quad
      const newRadius = RADIUS_60 + (eased * (RADIUS_100 - RADIUS_60));
      setRadius(newRadius);
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => {
      window.removeEventListener('scroll', handleScroll);
    };
  }, [introComplete]);

  return (
    <div
      className="fixed inset-0 -z-10"
      style={{
        backgroundColor: '#000000',
        backgroundRepeat: 'no-repeat',
        boxSizing: 'border-box',
        contain: 'content',
      }}
    >
      <StaticRadialGradient
        scale={2.24}
        offsetX={0}
        offsetY={0.38}
        radius={radius}
        focalDistance={2.3}
        focalAngle={360}
        falloff={1}
        mixing={0.26}
        distortionShift={0}
        distortionFreq={12}
        grainMixer={0.37}
        grainOverlay={0}
        colors={['#000000', '#006CFF', '#000000']}
        colorBack="#00000000"
        style={{
          height: '100vh',
          width: '100vw',
          position: 'absolute',
          top: '0',
          left: '0',
        }}
      />
    </div>
  );
}
