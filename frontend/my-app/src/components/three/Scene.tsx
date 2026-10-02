'use client';

import { Suspense } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import * as THREE from 'three';
import type { Device, RoomData, RoomLayout } from '@/lib/api/types';
import { HouseModel } from './HouseModel';

type SceneProps = {
  devices: Device[];
  rooms: RoomData[];
  roomLayout: RoomLayout[];
  selectedDeviceId: string | null;
  selectedRoomId: string | null;
  onSelectDevice: (deviceId: string) => void;
  onSelectRoom: (roomId: string) => void;
  onClearSelection: () => void;
};

export default function Scene({ devices, rooms, roomLayout, selectedDeviceId, selectedRoomId, onSelectDevice, onSelectRoom, onClearSelection }: SceneProps) {
  return (
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
          selectedDeviceId={selectedDeviceId}
          selectedRoomId={selectedRoomId}
          onSelectDevice={onSelectDevice}
          onSelectRoom={onSelectRoom}
        />
      </Suspense>
      <OrbitControls enablePan enableZoom enableRotate makeDefault />
    </Canvas>
  );
}