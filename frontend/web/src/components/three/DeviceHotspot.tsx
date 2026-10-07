import { Html } from '@react-three/drei';
import type { ThreeEvent } from '@react-three/fiber';
import { useState } from 'react';

type DeviceHotspotProps = {
  name: string;
  position: [number, number, number];
  on: boolean;
  online: boolean;
  selected: boolean;
  onClick: (event: ThreeEvent<MouseEvent>) => void;
};

export function DeviceHotspot({ name, position, on, online, selected, onClick }: DeviceHotspotProps) {
  const [hovered, setHovered] = useState(false);
  const color = !online ? '#d15f56' : on ? '#39a56f' : '#8a9890';

  return (
    <group position={position}>
      <mesh
        onClick={onClick}
        onPointerOver={(event) => {
          event.stopPropagation();
          setHovered(true);
        }}
        onPointerOut={() => setHovered(false)}
      >
        <sphereGeometry args={[selected ? 0.18 : 0.13, 20, 20]} />
        <meshStandardMaterial color={color} emissive={color} emissiveIntensity={selected ? 0.35 : 0.12} />
      </mesh>
      {selected && <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.1, 0]}>
        <ringGeometry args={[0.22, 0.27, 32]} />
        <meshBasicMaterial color="#1f6f5b" transparent opacity={0.9} />
      </mesh>}
      {(hovered || selected) && (
        <Html distanceFactor={9} zIndexRange={[10, 0]}>
          <div className="pointer-events-none -translate-x-1/2 -translate-y-10 whitespace-nowrap rounded-md bg-[#17221d] px-2 py-1 text-[10px] font-semibold text-white shadow-lg">
            {name}
          </div>
        </Html>
      )}
    </group>
  );
}