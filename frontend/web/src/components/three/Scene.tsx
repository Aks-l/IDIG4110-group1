'use client';

import { Component, Suspense, type ReactNode } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import * as THREE from 'three';
import type { Device, RoomConnection, RoomData, RoomLayout } from '@/lib/api/types';
import { HouseModel } from './HouseModel';

type SceneProps = {
  devices: Device[];
  rooms: RoomData[];
  roomLayout: RoomLayout[];
  connections: RoomConnection[];
  selectedDeviceId: string | null;
  selectedRoomId: string | null;
  onSelectDevice: (deviceId: string) => void;
  onSelectRoom: (roomId: string) => void;
  onClearSelection: () => void;
};

// Cached after the first check: creating a probe context is not free.
// three (r163+) requires a WebGL2 context, so WebGL1-only browsers count
// as unsupported — the renderer would throw on them anyway.
let webglProbeResult: boolean | null = null;

function probeWebgl(): boolean {
  if (typeof window === 'undefined') return false;
  try {
    const canvas = document.createElement('canvas');
    return !!canvas.getContext('webgl2');
  } catch {
    return false;
  }
}

function webglAvailable(): boolean {
  if (webglProbeResult === null) webglProbeResult = probeWebgl();
  return webglProbeResult;
}

function SceneFallback() {
  return (
    <div className="flex h-full min-h-[520px] flex-col items-center justify-center gap-3 p-8 text-center">
      <p className="text-base font-semibold text-[#17221d]">3D view unavailable</p>
      <p className="max-w-md text-sm text-[#68766d]">
        Your browser could not create a WebGL context, which the 3D model needs. Enable
        hardware acceleration in your browser settings (Chrome: Settings &rarr; System) or
        update your graphics drivers, then reload the page. The rest of the dashboard still
        works &mdash; use the room buttons below to inspect devices.
      </p>
    </div>
  );
}

// Renderer failures the probe cannot predict (shader or driver crashes at
// mount time) should degrade to the same fallback, not take the page down.
class SceneErrorBoundary extends Component<{ children: ReactNode }, { failed: boolean }> {
  state = { failed: false };

  static getDerivedStateFromError() {
    return { failed: true };
  }

  render() {
    return this.state.failed ? <SceneFallback /> : this.props.children;
  }
}

export default function Scene({ devices, rooms, roomLayout, connections, selectedDeviceId, selectedRoomId, onSelectDevice, onSelectRoom, onClearSelection }: SceneProps) {
  if (!webglAvailable()) return <SceneFallback />;
  return (
    <SceneErrorBoundary>
    <Canvas
      camera={{ position: [8, 8, 10], fov: 42 }}
      shadows={{ type: THREE.PCFShadowMap }}
      onPointerMissed={onClearSelection}
      dpr={[1, 2]}
    >
      <color attach="background" args={['#edf3ee']} />
      <ambientLight intensity={1.8} />
      <directionalLight position={[5, 10, 5]} intensity={2.5} castShadow />
      <gridHelper args={[16, 16, '#c1d1c4', '#dce7de']} position={[0, -0.02, 0]} />
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.06, 0]} onClick={onClearSelection}>
        <planeGeometry args={[16, 16]} />
        <meshStandardMaterial color="#f8faf7" />
      </mesh>
      <Suspense fallback={null}>
        <HouseModel
          devices={devices}
          rooms={rooms}
          roomLayout={roomLayout}
          connections={connections}
          selectedDeviceId={selectedDeviceId}
          selectedRoomId={selectedRoomId}
          onSelectDevice={onSelectDevice}
          onSelectRoom={onSelectRoom}
        />
      </Suspense>
      <OrbitControls enablePan enableZoom enableRotate makeDefault />
    </Canvas>
    </SceneErrorBoundary>
  );
}