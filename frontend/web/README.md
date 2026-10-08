This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Getting Started

### Run with Docker Compose

From this directory, build and start the frontend with:

```bash
docker compose up --build
```

Open [http://localhost:3000](http://localhost:3000). Compose builds with the
same defaults as `npm run dev` with `.env.local`: the pages call the
placeholder API routes (`NEXT_PUBLIC_DATA_SOURCE=api`), and `/automations`
plus `/events` are served from the rules-engine (hybrid mode). The container
reaches the backend stack's published ports through `host.docker.internal`,
so bring the backend up first from the repository root:

```bash
docker compose up -d   # repository root
```

Without the backend stack the automations and events pages show 502 errors;
everything else keeps serving fixtures. The defaults live in
`docker-compose.yml` and can be overridden from the shell (an empty variable
falls back to the built-in default):

```bash
# back to mock mode without the placeholder routes
NEXT_PUBLIC_DATA_SOURCE=mock docker compose up --build

# also serve /rooms, /devices and /3d/rooms from twin-core (opt-in until
# the scaffold branch merges; PLACEHOLDER_TWIN_PATHS defaults to these)
PLACEHOLDER_TWIN_URL=http://host.docker.internal:8084 docker compose up -d
```

`NEXT_PUBLIC_*` variables are baked into the browser bundle at build time —
after changing them, rebuild with `docker compose up --build`. The
`PLACEHOLDER_*` variables are read server-side at request time and only need
the container to be recreated (`docker compose up -d`).

To connect a real backend instead, either point the browser at it directly
(`NEXT_PUBLIC_API_URL=http://<backend>/api`), or keep the browser on the
placeholder routes and set `PLACEHOLDER_API_TARGET=http://<backend>/api` — the
placeholder routes then forward every request there. Either way no code
changes are needed. `docs/frontend/api.md` documents the contract and both
hybrid modes per endpoint.

Stop the container with `docker compose down`.

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
