# NEXUS Digital Twin UI

A desktop-first industrial Digital Twin console built with Next.js, TypeScript, Tailwind CSS, React Three Fiber, Drei, Recharts, Zustand, Zod-ready domain types, and Lucide icons.

## Run locally

```bash
cd src/ui
npm install
npm run dev
```

Open http://localhost:3000.

## Production checks

```bash
npm run lint
npm run build
```

The mock data boundary lives in `src/data/mockTwinData.ts`. Replace that module with REST/WebSocket/MQTT-backed service implementations without changing the UI contracts. Global selection and shell state live in `src/stores/useTwinStore.ts`.
