import { create } from "zustand";
import { rooms as initialRooms, sensors as initialSensors } from "@/data/mockTwinData";
import { Room, Sensor } from "@/types";

interface TwinStore {
  rooms: Room[];
  sensors: Sensor[];
  selectedRoomId: string;
  selectedSensorId: string | null;
  isSidebarOpen: boolean;
  isDetailsOpen: boolean;
  isDraggingSensor: boolean;
  activePage: string;
  setSelectedRoom: (id: string) => void;
  setSelectedSensor: (id: string | null) => void;
  setSidebarOpen: (open: boolean) => void;
  setDetailsOpen: (open: boolean) => void;
  setDraggingSensor: (dragging: boolean) => void;
  setActivePage: (page: string) => void;
  addSensor: (sensor: Sensor) => void;
  updateSensor: (id: string, updates: Partial<Sensor>) => void;
  removeSensor: (id: string) => void;
  getSensor: (id: string) => Sensor | undefined;
  getSensorsByRoom: (roomId: string) => Sensor[];
}

export const useTwinStore = create<TwinStore>((set, get) => ({
  rooms: initialRooms,
  sensors: initialSensors,
  selectedRoomId: "living-room",
  selectedSensorId: null,
  isSidebarOpen: true,
  isDetailsOpen: true,
  isDraggingSensor: false,
  activePage: "Overview",
  setSelectedRoom: (id) => set({ selectedRoomId: id, selectedSensorId: null, isDetailsOpen: true }),
  setSelectedSensor: (id) => set((state) => {
    const sensor = id ? state.sensors.find((item) => item.id === id) : undefined;
    return { selectedSensorId: id, selectedRoomId: sensor?.roomId ?? state.selectedRoomId, isDetailsOpen: true };
  }),
  setSidebarOpen: (open) => set({ isSidebarOpen: open }),
  setDetailsOpen: (open) => set({ isDetailsOpen: open }),
  setDraggingSensor: (dragging) => set({ isDraggingSensor: dragging }),
  setActivePage: (page) => set({ activePage: page }),
  addSensor: (sensor) => set((state) => ({ sensors: [...state.sensors, sensor], selectedSensorId: sensor.id, isDetailsOpen: true })),
  updateSensor: (id, updates) => set((state) => ({ sensors: state.sensors.map((sensor) => sensor.id === id ? { ...sensor, ...updates } : sensor) })),
  removeSensor: (id) => set((state) => ({ sensors: state.sensors.filter((sensor) => sensor.id !== id), selectedSensorId: state.selectedSensorId === id ? null : state.selectedSensorId })),
  getSensor: (id) => get().sensors.find((sensor) => sensor.id === id),
  getSensorsByRoom: (roomId) => get().sensors.filter((sensor) => sensor.roomId === roomId),
}));
