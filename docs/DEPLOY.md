# Deploying SayCanto for free

- **API** (Go + SQLite): Google Cloud Run, built from this repo's root `Dockerfile` by Cloud Build on every push to `main`.
- **Website** (Next.js): Vercel Hobby plan, from the `frontend/` folder.

Free-tier limits change; check the current [Cloud Run](https://cloud.google.com/run/pricing),
[Cloud Build](https://cloud.google.com/build/pricing), [Artifact Registry](https://cloud.google.com/artifact-registry/pricing)
and [Vercel Hobby](https://vercel.com/docs/plans/hobby) pages. Vercel Hobby is for non-commercial use only.

## 0. The dictionary file

The image downloads the full dictionary from a GitHub Release and checks its SHA-256 (`DICT_URL` and
`DICT_SHA256` in the `Dockerfile`). Current release: tag `data-2026-11-27`, asset `dictionary.db.gz`.

To publish a new dictionary after changing the data:
1. Rebuild it with the full importer command in `data/README.md`.
2. `gzip -c -9 -n backend/dictionary.db > dist/dictionary.db.gz` and `sha256sum dist/dictionary.db.gz`.
3. Create a new release (new tag) with `dist/dictionary.db.gz` and `data/CREDITS.md` attached.
4. Update `DICT_URL` and `DICT_SHA256` in the `Dockerfile`, commit and push. Cloud Run redeploys.

## 1. Google Cloud project (one time)

1. Sign in at https://console.cloud.google.com and create a project, e.g. `saycanto`.
2. **Billing**: link a billing account (a card is required even for the free tier).
3. **Budget alert**: Billing → Budgets & alerts → create a budget of e.g. US$1 with email alerts.
   An alert does not stop spending; the instance limit in step 2 below keeps costs bounded.

## 2. Deploy the API on Cloud Run

1. Open **Cloud Run** → **Deploy container** → **Service**.
2. Choose **Continuously deploy from a repository**, pick **Cloud Build**, then **Set up with Cloud Build**:
   - Repository provider: **GitHub**; click **Authenticate** in the panel and approve the Google Cloud Build
     GitHub app for this repository only; pick `kaho-dev/saycanto` and tick the data-transfer consent.
   - Branch: `^main$`.
   - Build type: **Dockerfile**, source location `/Dockerfile`.
3. Service settings:
   - **Service name**: `saycanto-api`.
   - **Region**: `europe-west1` (Belgium) is what the live service uses; any Tier 1 region works.
   - **Authentication**: allow public (unauthenticated) access. The API is read-only.
   - **Billing**: request-based (CPU only while handling requests).
   - **Scaling**: minimum instances **0**, maximum instances **2**.
   - **Container**: port `8080`, memory `512 MiB`, CPU `1`.
   - **Variables**: none needed yet; the image defaults `CORS_ORIGIN` to `*` (set it in step 4).
   - **Security → Service account**: *Default compute service account* (the form will not submit without one).
4. Create. When it is ready, copy the service URL (`https://saycanto-…run.app`) and check
   `https://<service-url>/api/health` returns `{"status":"ok"}`.
5. **Artifact Registry** keeps every built image (~100 MB each). Add a cleanup policy that keeps
   only the 2 most recent versions so storage stays inside the free 0.5 GB.

## Live setup (27 Nov 2026)

| | |
|---|---|
| Website | https://saycanto.vercel.app (Vercel team "MarshalBuilder", Hobby; project `saycanto`, root `frontend`) |
| API | https://saycanto-api-417998879746.europe-west1.run.app (Cloud Run service `saycanto-api`, project `saycanto-516012`, `europe-west1`) |
| API settings | 0–2 instances, request-based billing, 512 MiB, public access, default compute service account, `CORS_ORIGIN=https://saycanto.vercel.app` |
| Billing | Billing account "My Billing Account 2", €1/month budget alert |
| Image storage | Artifact Registry `cloud-run-source-deploy`: cleanup policies `delete-older-than-1-day` + `keep-2-most-recent` |
| Deploys | Every push to `main` rebuilds both (Cloud Build trigger for the API, Vercel Git integration for the site) |

Note: with `CORS_ORIGIN` set to the production address, Vercel **preview** deployments (other URLs) cannot
call the API from the browser; their server-rendered word pages still work.

## 3. Deploy the website on Vercel

1. Sign in at https://vercel.com with GitHub → **Add New… → Project** → import `kaho-dev/saycanto`.
2. **Root Directory**: `frontend` (framework is detected as Next.js).
3. **Environment variable**: `NEXT_PUBLIC_API_URL` = the Cloud Run URL from step 2 (no trailing slash).
4. Deploy, then open the site and search for "thank you".

## 4. Lock the API to the website

In Cloud Run → the service → **Containers** tab → Variables & Secrets → **Add variable**: `CORS_ORIGIN` =
the website address, e.g. `https://saycanto.vercel.app`, then **View diff & redeploy** → Deploy changes.
Environment variables survive later automatic deploys from Cloud Build. Browsers on other sites can then no
longer call the API directly.

## Checks after each deploy

- `https://<service-url>/api/search?q=usually&limit=3` lists 通常 first.
- The website loads starter words, search works, and the footer credits link to `data/CREDITS.md`.
- CI's `docker` job builds the same image and smoke-tests it on every push.
