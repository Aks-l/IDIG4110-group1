import { useMemo } from 'react';
import * as THREE from 'three';
import { useGLTF } from '@react-three/drei';
import type { ThreeEvent } from '@react-three/fiber';
import type { Device, RoomData, RoomLayout } from '@/lib/api/types';
import { DeviceHotspot } from './DeviceHotspot';

type HouseModelProps = {
  rooms: RoomData[];
  roomLayout: RoomLayout[];
  devices: Device[];
  selectedRoomId: string | null;
  selectedDeviceId: string | null;
  onSelectRoom: (roomId: string) => void;
  onSelectDevice: (deviceId: string) => void;
};

const roomColors = ['#a8cdb8', '#c9d8bb', '#b6cad0', '#d8c7a7'];
const USE_REAL_MODEL = false;

const connections = [
  { position: [0.075, 0.3, 1.75] as [number, number, number], size: [1.37, 0.6, 0.85] as [number, number, number] },
  { position: [-2.85, 0.3, -0.325] as [number, number, number], size: [1.05, 0.6, 0.87] as [number, number, number] },
  { position: [3.2, 0.3, -0.3] as [number, number, number], size: [0.95, 0.6, 1.42] as [number, number, number] },
];

function RealHouseModel() {
  const { scene } = useGLTF('/models/house.glb');
  return <primitive object={scene} />;
}

export function HouseModel({ rooms, roomLayout, devices, selectedRoomId, selectedDeviceId, onSelectRoom, onSelectDevice }: HouseModelProps) {
  const roomMap = useMemo(() => new Map(rooms.map((room) => [room.id, room])), [rooms]);
  const layoutMap = useMemo(() => new Map(roomLayout.map((layout) => [layout.roomId, layout])), [roomLayout]);

  // TODO: Replace this block-out with useGLTF('/models/house.glb') when the real model is available.
  const getDevicePosition = (device: Device, index: number): [number, number, number] => {
    const layout = device.roomId ? layoutMap.get(device.roomId) : undefined;
    if (device.position) return device.position;
    const center = layout?.position ?? [0, 0.9, 0];
    return [center[0] + ((index % 3) - 1) * 0.8, 0.9, center[2] + (index % 2) * 0.65 - 0.3];
  };

  return (
    <group>
      {USE_REAL_MODEL && <RealHouseModel />}
      <group>
        {connections.map((connection, index) => (
          <mesh key={`connection-${index}`} position={connection.position}>
            <boxGeometry args={connection.size} />
            <meshStandardMaterial color="#d8e5da" />
          </mesh>
        ))}
      </group>
      {roomLayout.map((layout, index) => {
        const room = roomMap.get(layout.roomId);
        if (!room) return null;
        const selected = selectedRoomId === room.id;
        return (
          <group key={room.id} position={layout.position}>
            <mesh
              onClick={(event: ThreeEvent<MouseEvent>) => {
                event.stopPropagation();
                onSelectRoom(room.id);
              }}
              onPointerOver={(event) => { event.stopPropagation(); }}
            >
              <boxGeometry args={[layout.size[0], layout.size[1], layout.size[2]]} />
              <meshStandardMaterial color={selected ? '#5eaa87' : roomColors[index % roomColors.length]} transparent opacity={0.92} />
            </mesh>
            <lineSegments position={[0, layout.size[1] / 2 + 0.01, 0]}>
              <edgesGeometry args={[new THREE.BoxGeometry(layout.size[0], layout.size[1], layout.size[2])]} />
              <lineBasicMaterial color={selected ? '#145442' : '#78917f'} />
            </lineSegments>
          </group>
        );
      })}
      {devices.map((device, index) => (
        <DeviceHotspot
          key={device.id}
          name={device.name}
          position={getDevicePosition(device, index)}
          on={device.on}
          online={true}
          selected={selectedDeviceId === device.id}
          onClick={(event) => {
            event.stopPropagation();
            onSelectDevice(device.id);
          }}
        />
      ))}
    </group>
  );
}