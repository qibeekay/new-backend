# Render Deployment & Keep-Alive Guide

This guide walks you through deploying the `portfolio-api` backend to Render's free tier and ensuring the service stays awake without entering the 15-minute inactivity sleep.

---

## 1. How the Anti-Sleep (Keep-Alive) Mechanism Works

Render free tier instances automatically spin down to save resources after **15 minutes of zero inbound HTTP traffic**.

To prevent this, two complementary mechanisms are set up:

### Mechanism A: Built-in Go Self-Pinger (Already Active in Code)
- Render automatically injects the environment variable `RENDER_EXTERNAL_URL` (e.g. `https://portfolio-api.onrender.com`).
- The backend detects this variable on startup and starts a background goroutine in `main.go`.
- Every **10 minutes**, it sends an HTTP `GET` request to its own public URL:
  ```
  https://<your-service>.onrender.com/healthz
  ```
- Because this request originates and routes through Render's external router, Render registers it as active incoming web traffic, resetting the 15-minute idle timer.

### Mechanism B: External Uptime Monitor (Recommended 1-Minute Setup)
If Render ever restarts or experiences a cold boot after maintenance, an external pinger guarantees the service is woken up immediately:
1. Create a free account at [UptimeRobot.com](https://uptimerobot.com) (or [cron-job.org](https://cron-job.org)).
2. Click **Add New Monitor**.
3. Select **Monitor Type:** `HTTP(s)`.
4. Enter **Friendly Name:** `Portfolio API`.
5. Enter **URL:** `https://<your-render-service-url>.onrender.com/healthz`.
6. Set **Monitoring Interval:** `5 minutes` or `10 minutes`.
7. Click **Create Monitor**.

---

## 2. Deploying to Render (Step-by-Step)

### Option 1: Via Render Dashboard (Recommended)

1. Push your `portfolio-api` repository to GitHub or GitLab.
2. Go to your [Render Dashboard](https://dashboard.render.com/) and click **New +** -> **Web Service**.
3. Connect your Git repository.
4. Fill in the service settings:
   - **Name:** `portfolio-api` (or your preferred name)
   - **Region:** `Oregon (US West)` or closest to your Neon database (`Ohio (US East)`)
   - **Branch:** `main`
   - **Root Directory:** *(leave blank if root)*
   - **Runtime:** `Go`
   - **Build Command:**
     ```bash
     go run ./migrate up && go build -o bin/api ./cmd/api
     ```
     *(This automatically runs database migrations against Neon on every deploy before compiling)*
   - **Start Command:**
     ```bash
     ./bin/api
     ```
   - **Instance Type:** `Free`

5. Under **Advanced** -> **Health Check Path**, enter:
   ```
   /healthz
   ```

---

## 3. Environment Variables to Set on Render

In your Render Web Service dashboard under the **Environment** tab, add the following variables:

| Key | Value / Source | Description |
| :--- | :--- | :--- |
| `PORT` | `10000` | Render standard port |
| `DATABASE_URL` | `postgresql://<user>:<password>@<neon-host>/neondb?sslmode=require` | Your Neon connection string from Neon Console |
| `JWT_SECRET` | `<your-32-character-secret>` | Random 32+ character string for JWT tokens |
| `ADMIN_EMAIL` | `admin@example.com` | Admin login email |
| `ADMIN_PASSWORD_HASH` | `<bcrypt-hash>` | Generated bcrypt hash (generate with `go run ./cmd/hashpw <password>`) |
| `ALLOWED_ORIGINS` | `http://localhost:5173,https://qibeekay.vercel.app` | Comma-separated frontend URLs allowed by CORS |
| `CLOUDINARY_CLOUD_NAME` | `<your-cloud-name>` | Cloudinary Cloud Name |
| `CLOUDINARY_API_KEY` | `<your-api-key>` | Cloudinary API Key |
| `CLOUDINARY_API_SECRET` | `<your-api-secret>` | Cloudinary API Secret |
| `CLOUDINARY_FOLDER` | `portfolio` | Cloudinary media folder |
| `MAX_UPLOAD_MB` | `5` | Maximum image upload size |

> [!NOTE]
> `RENDER_EXTERNAL_URL` is automatically provided by Render at runtime. You do not need to manually configure `KEEP_ALIVE_URL` unless you want to override it.

---

## 4. Verification After Deployment

Once the deployment finishes:

1. **Verify Health:**
   ```bash
   curl https://<your-service>.onrender.com/healthz
   # Expected response: {"ok": true}
   ```
2. **Access Swagger UI:**
   Open `https://<your-service>.onrender.com/swagger` in your browser to test endpoints directly.
3. **Check Render Logs:**
   In the Render log console, you will see:
   ```
   keep-alive: self-pinger enabled for https://<your-service>.onrender.com/healthz (interval: 10m)
   listening on :10000
   ```
   Every 10 minutes, you will see:
   ```
   keep-alive: ping successful (status 200)
   ```
