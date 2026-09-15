"use client";

import { Environment, Html, OrbitControls, Text } from "@react-three/drei";
import { Canvas, ThreeEvent } from "@react-three/fiber";
import type { OrbitControls as OrbitControlsImpl } from "three-stdlib";
import { useEffect, useRef, useState } from "react";
import { useTwinStore } from "@/stores/useTwinStore";
import { Room, RoomStatus, Sensor } from "@/types";

const roomColor: Record<RoomStatus, string> = { normal: "#41c695", warning: "#f4b740", critical: "#f05d5e" };
const sensorColor: Record<Sensor["status"], string> = { normal: "#5bc0eb", warning: "#f4b740", critical: "#f05d5e", offline: "#687487" };

function RoomTile({ room }: { room: Room }) {
  const [hovered, setHovered] = useState(false);
  const selectedRoomId = useTwinStore((state) => state.selectedRoomId);
  const setSelectedRoom = useTwinStore((state) => state.setSelectedRoom);
  const selected = selectedRoomId === room.id;
  const color = roomColor[room.status];
  const handleClick = (event: ThreeEvent<MouseEvent>) => { event.stopPropagation(); setSelectedRoom(room.id); };
  return <group position={room.position} onClick={handleClick} onPointerOver={() => setHovered(true)} onPointerOut={() => setHovered(false)}>
    <mesh receiveShadow><boxGeometry args={[room.size[0], room.size[1], room.size[2]]} /><meshStandardMaterial color={selected || hovered ? "#294354" : "#182833"} metalness={0.25} roughness={0.7} /></mesh>
    <mesh position={[0, 0.13, 0]}><boxGeometry args={[room.size[0] - 0.12, 0.02, room.size[2] - 0.12]} /><meshStandardMaterial color="#213744" roughness={0.95} /></mesh>
    <mesh position={[0, 0.18, -room.size[2] / 2 + 0.06]}><boxGeometry args={[room.size[0] - 0.18, 0.03, 0.08]} /><meshStandardMaterial color="#b7c7cf" metalness={0.2} /></mesh>
    {selected && <mesh position={[0, 0.2, 0]}><boxGeometry args={[room.size[0] - 0.08, 0.03, room.size[2] - 0.08]} /><meshBasicMaterial color={color} wireframe transparent opacity={0.75} /></mesh>}
    <mesh position={[room.size[0] / 2 - 0.25, 0.23, room.size[2] / 2 - 0.25]}><sphereGeometry args={[0.08, 12, 12]} /><meshStandardMaterial color={color} emissive={color} emissiveIntensity={0.18} /></mesh>
    <Text position={[0, 0.28, 0]} rotation={[-Math.PI / 2, 0, 0]} fontSize={0.16} color="#dbe7ef" anchorX="center" anchorY="middle">{room.name}</Text>
    {hovered && <Html position={[0, 0.65, 0]} center><div className="border border-slate-600 bg-slate-950/95 px-2 py-1 text-[10px] text-slate-200 shadow-xl">{room.temperature}°C · {room.co2} ppm</div></Html>}
  </group>;
}

function SensorMarker({ sensor }: { sensor: Sensor }) {
  const [hovered, setHovered] = useState(false);
  const dragging = useRef(false);
  const selectedSensorId = useTwinStore((state) => state.selectedSensorId);
  const rooms = useTwinStore((state) => state.rooms);
  const setSelectedSensor = useTwinStore((state) => state.setSelectedSensor);
  const setDraggingSensor = useTwinStore((state) => state.setDraggingSensor);
  const updateSensor = useTwinStore((state) => state.updateSensor);
  const selected = selectedSensorId === sensor.id;
  const color = sensorColor[sensor.status];
  const room = rooms.find((item) => item.id === sensor.roomId);
  const handlePointerDown = (event: ThreeEvent<PointerEvent>) => {
    event.stopPropagation();
    dragging.current = true;
    (event.target as unknown as { setPointerCapture: (pointerId: number) => void }).setPointerCapture(event.pointerId);
    setDraggingSensor(true);
    setSelectedSensor(sensor.id);
  };
  const handlePointerMove = (event: ThreeEvent<PointerEvent>) => {
    if (!dragging.current || !room) return;
    event.stopPropagation();
    const padding = 0.22;
    const x = Math.max(room.position[0] - room.size[0] / 2 + padding, Math.min(room.position[0] + room.size[0] / 2 - padding, event.point.x));
    const z = Math.max(room.position[2] - room.size[2] / 2 + padding, Math.min(room.position[2] + room.size[2] / 2 - padding, event.point.z));
    updateSensor(sensor.id, { position: [x, sensor.position[1], z] });
  };
  const handlePointerUp = (event: ThreeEvent<PointerEvent>) => {
    event.stopPropagation();
    dragging.current = false;
    (event.target as unknown as { releasePointerCapture: (pointerId: number) => void }).releasePointerCapture(event.pointerId);
    setDraggingSensor(false);
  };
  const scale = selected ? 1.22 : 1;
  const material = <meshStandardMaterial color={color} emissive={color} emissiveIntensity={selected ? 0.65 : 0.16} metalness={0.35} roughness={0.4} />;
  const glyph = sensor.type === "motion" ? <><mesh rotation={[Math.PI / 2, 0, 0]}><ringGeometry args={[0.11, 0.15, 20]} />{material}</mesh><mesh rotation={[Math.PI / 2, 0, 0]}><ringGeometry args={[0.24, 0.27, 20]} /><meshBasicMaterial color={color} transparent opacity={selected ? 0.75 : 0.35} /></mesh></> : sensor.type === "light" ? <><mesh position={[0, 0.1, 0]}><coneGeometry args={[0.16, 0.2, 8]} />{material}</mesh><mesh position={[0, -0.07, 0]}><cylinderGeometry args={[0.06, 0.06, 0.12, 12]} />{material}</mesh></> : sensor.type === "door" || sensor.type === "window" ? <><mesh><boxGeometry args={[0.24, 0.08, 0.16]} />{material}</mesh><mesh position={[0.16, 0, 0]}><boxGeometry args={[0.08, 0.14, 0.16]} />{material}</mesh></> : sensor.type === "water-leak" ? <><mesh rotation={[-Math.PI / 2, 0, 0]}><circleGeometry args={[0.2, 16]} />{material}</mesh><mesh position={[0, 0.04, 0]}><sphereGeometry args={[0.055, 12, 8]} />{material}</mesh></> : sensor.type === "smoke" || sensor.type === "co2" || sensor.type === "air-quality" ? <><mesh><cylinderGeometry args={[0.13, 0.13, 0.16, 12]} />{material}</mesh><mesh position={[0, 0.12, 0]}><cylinderGeometry args={[0.07, 0.07, 0.04, 12]} />{material}</mesh></> : <><mesh><cylinderGeometry args={[0.075, 0.11, 0.25, 12]} />{material}</mesh><mesh position={[0, 0.17, 0]}><sphereGeometry args={[0.09, 12, 8]} />{material}</mesh></>;
  return <group position={sensor.position} scale={scale} onPointerDown={handlePointerDown} onPointerMove={handlePointerMove} onPointerUp={handlePointerUp} onPointerCancel={handlePointerUp} onPointerOver={() => setHovered(true)} onPointerOut={() => setHovered(false)}>
    {glyph}
    <mesh rotation={[Math.PI / 2, 0, 0]}><ringGeometry args={[0.18, 0.2, 20]} /><meshBasicMaterial color={color} transparent opacity={selected ? 0.8 : 0.28} /></mesh>
    {(hovered || selected) && <Html position={[0, 0.28, 0]} center><div className="border border-slate-600 bg-slate-950/95 px-2 py-1 text-[10px] text-slate-200 shadow-xl">{sensor.name}<span className="ml-2 text-cyan-300">{sensor.value} {sensor.unit}</span></div></Html>}
  </group>;
} 

function HouseModel({ focusRoom }: { focusRoom: boolean }) {
  const rooms = useTwinStore((state) => state.rooms);
  const sensors = useTwinStore((state) => state.sensors);
  const selectedRoomId = useTwinStore((state) => state.selectedRoomId);
  const visibleRooms = focusRoom ? rooms.filter((room) => room.id === selectedRoomId) : rooms;
  const visibleSensors = focusRoom ? sensors.filter((sensor) => sensor.roomId === selectedRoomId) : sensors;
  return <group position={[0, -0.15, 0]}>
    <mesh rotation={[-Math.PI / 2, 0, 0]} receiveShadow><planeGeometry args={[11, 8]} /><meshStandardMaterial color="#101a22" roughness={0.95} /></mesh>
    <gridHelper args={[11, 22, "#314252", "#1a2935"]} position={[0, 0.02, 0]} />
    {visibleRooms.map((room) => <RoomTile key={room.id} room={room} />)}
    {visibleSensors.map((sensor) => <SensorMarker key={sensor.id} sensor={sensor} />)}
    {!focusRoom && <><mesh position={[-5.55, 1.1, 0]} castShadow><boxGeometry args={[0.12, 2.4, 7.3]} /><meshStandardMaterial color="#314552" /></mesh><mesh position={[5.55, 1.1, 0]} castShadow><boxGeometry args={[0.12, 2.4, 7.3]} /><meshStandardMaterial color="#314552" /></mesh><mesh position={[0, 1.1, -3.7]} castShadow><boxGeometry args={[11.2, 2.4, 0.12]} /><meshStandardMaterial color="#314552" /></mesh></>}
  </group>;
}

export default function TwinScene({ compact = false, focusRoom = false }: { compact?: boolean; focusRoom?: boolean }) {
  const isDraggingSensor = useTwinStore((state) => state.isDraggingSensor);
  const selectedRoomId = useTwinStore((state) => state.selectedRoomId);
  const selectedRoom = useTwinStore((state) => state.rooms.find((room) => room.id === selectedRoomId));
  const controlsRef = useRef<OrbitControlsImpl | null>(null);
  const selectedRoomX = selectedRoom?.position[0];
  const selectedRoomZ = selectedRoom?.position[2];
  const cameraPosition: [number, number, number] = focusRoom && selectedRoom ? [selectedRoom.position[0] + 4.5, 5.5, selectedRoom.position[2] + 4.5] : [8, 8, 9];
  useEffect(() => {
    if (!controlsRef.current) return;
    controlsRef.current.target.set(focusRoom && selectedRoomX !== undefined ? selectedRoomX : 0, focusRoom ? 0.35 : 0.4, focusRoom && selectedRoomZ !== undefined ? selectedRoomZ : 0);
    controlsRef.current.update();
  }, [focusRoom, selectedRoomId, selectedRoomX, selectedRoomZ]);
  return <div className={compact ? "h-full min-h-[300px] w-full" : "h-full min-h-[540px] w-full"}><Canvas key={`${focusRoom}-${selectedRoomId}`} shadows camera={{ position: cameraPosition, fov: focusRoom ? 46 : 42 }} dpr={[1, 1.5]}><color attach="background" args={["#0c141b"]} /><fog attach="fog" args={["#0c141b", 14, 28]} /><ambientLight intensity={0.55} /><directionalLight position={[5, 8, 5]} intensity={1.4} castShadow shadow-mapSize={[1024, 1024]} /><Environment preset="city" background={false} /><HouseModel focusRoom={focusRoom} /><OrbitControls ref={controlsRef} makeDefault enabled={!isDraggingSensor} enablePan screenSpacePanning panSpeed={1.2} rotateSpeed={0.8} zoomSpeed={1.2} minDistance={focusRoom ? 1.4 : 3} maxDistance={focusRoom ? 16 : 30} minPolarAngle={0.12} maxPolarAngle={Math.PI * 0.92} /></Canvas></div>;
}
