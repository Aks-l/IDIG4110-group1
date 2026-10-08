This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Getting Started

### Run with Docker Compose

From this directory, build and start the frontend with:

```bash
docker compose up --build
```

Open [http://localhost:3000](http://localhost:3000). The default Compose setup
uses mock data. The app also ships placeholder API routes (`src/app/api`) that
serve the frontend contract with fixture data, so `NEXT_PUBLIC_DATA_SOURCE=api`
works without a backend:

```env
NEXT_PUBLIC_DATA_SOURCE=api
NEXT_PUBLIC_API_URL=http://localhost:3000/api
```

To connect a real backend, either point the browser at it directly
(`NEXT_PUBLIC_API_URL=http://<backend>/api`), or keep the browser on the
placeholder routes and set `PLACEHOLDER_API_TARGET=http://<backend>/api` — the
placeholder routes then forward every request there. Either way no code changes
are needed. See `docs/frontend/api.md` for the full contract.

The endpoints the backend already implements can also be served from it while
the rest stay on fixtures (hybrid mode): set
`PLACEHOLDER_BACKEND_URL=http://localhost:8084` and
`PLACEHOLDER_BACKEND_PATHS=/automations,/events`, and `/automations` plus
`/events` are fetched from the api-gateway and translated onto the frontend
contract. `docs/frontend/api.md` documents the per-endpoint details.

Then rebuild the image:

```bash
docker compose up --build
```

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
