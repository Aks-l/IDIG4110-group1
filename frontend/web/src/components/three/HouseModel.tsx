import { useMemo } from 'react';
import * as THREE from 'three';
import { useGLTF } from '@react-three/drei';
import type { ThreeEvent } from '@react-three/fiber';
import type { Device, RoomConnection, RoomData, RoomLayout } from '@/lib/api/types';
import { DeviceHotspot } from './DeviceHotspot';

type HouseModelProps = {
  rooms: RoomData[];
  roomLayout: RoomLayout[];
  connections: RoomConnection[];
  devices: Device[];
  selectedRoomId: string | null;
  selectedDeviceId: string | null;
  onSelectRoom: (roomId: string) => void;
  onSelectDevice: (deviceId: string) => void;
};

const roomColors = ['#a8cdb8', '#c9d8bb', '#b6cad0', '#d8c7a7'];
const USE_REAL_MODEL = false;

function RealHouseModel() {
  const { scene } = useGLTF('/models/house.glb');
  return <primitive object={scene} />;
}

export function HouseModel({ rooms, roomLayout, connections, devices, selectedRoomId, selectedDeviceId, onSelectRoom, onSelectDevice }: HouseModelProps) {
  const roomMap = useMemo(() => new Map(rooms.map((room) => [room.id, room])), [rooms]);
  const layoutMap = useMemo(() => new Map(roomLayout.map((layout) => [layout.roomId, layout])), [roomLayout]);

  // TODO: Replace this block-out with useGLTF('/models/house.glb') when the real model is available.
  const getDevicePosition = (device: Device, index: number): [number, number, number] => {
    if (device.position) return device.position;
    const layout = device.roomId ? layoutMap.get(device.roomId) : undefined;
    if (!layout) return [((index % 3) - 1) * 0.8, 0.9, (index % 2) * 0.65 - 0.3];
    // Hotspots float just above the room box so they stay visible over solid
    // room volumes (0.3 above the mock slabs keeps the old 0.9).
    const top = layout.position[1] + layout.size[1] / 2;
    return [
      layout.position[0] + ((index % 3) - 1) * 0.8,
      top + 0.3,
      layout.position[2] + (index % 2) * 0.65 - 0.3,
    ];
  };

  return (
    <group>
      {USE_REAL_MODEL && <RealHouseModel />}
      <group>
        {connections.map((connection) => (
          <mesh
            key={`connection-${connection.fromRoomId}-${connection.toRoomId}`}
            position={connection.position}
          >
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