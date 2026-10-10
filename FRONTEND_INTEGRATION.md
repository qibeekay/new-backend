# Portfolio API Integration Specification

This document provides the complete API contract and integration specification for consuming the backend endpoints.

---

## 1. General Information

- **Base URL (Local Development):** `http://localhost:8080`
- **API Prefix:** `/api`
- **Interactive Swagger UI:** `http://localhost:8080/swagger`
- **OpenAPI JSON Spec:** `http://localhost:8080/swagger/doc.json`
- **CORS Allowed Origins:** Configured via `ALLOWED_ORIGINS` (defaults to `http://localhost:5173`, `http://localhost:3000`)
- **Default Admin Development Credentials:**
  - Email: `admin@example.com`
  - Password: `admin123`

---

## 2. Authentication & Headers

Protected admin endpoints require a JSON Web Token (JWT) in the `Authorization` header:

```http
Authorization: Bearer <YOUR_JWT_TOKEN>
```

- When the `Authorization` header is missing, malformed, or the token is expired/invalid, the API returns `401 Unauthorized`.
- Non-admin endpoints (public project listings) do not require the `Authorization` header.

---

## 3. Standard Error Formats

### Standard Error (HTTP 400, 401, 404, 413, 415, 500)
```json
{
  "error": "human-readable error message"
}
```

### Validation Error (HTTP 422 Unprocessable Entity)
Returned when payload validation fails. Provides a key-value map of invalid field names to error messages:
```json
{
  "errors": {
    "title": "Give the project a name.",
    "year": "Use a four-digit year.",
    "category": "Must be creative, functional or systems.",
    "status": "Must be completed, ongoing, ideating or paused.",
    "description": "Write at least a couple of sentences."
  }
}
```

---

## 4. Endpoints Specification

### 4.1. Health Check
Checks backend service availability.

- **Method:** `GET`
- **Endpoint:** `/healthz`
- **Auth Required:** No

#### Response:
- **Status:** `200 OK`
```json
{
  "ok": true
}
```

---

### 4.2. Admin Login
Authenticates admin credentials and returns a JWT token.

- **Method:** `POST`
- **Endpoint:** `/api/admin/login`
- **Auth Required:** No
- **Headers:** `Content-Type: application/json`

#### Request Payload:
| Field | Type | Required | Description | Example |
| :--- | :--- | :--- | :--- | :--- |
| `email` | `string` | Yes | Registered admin email | `"admin@example.com"` |
| `password` | `string` | Yes | Admin password | `"admin123"` |

```json
{
  "email": "admin@example.com",
  "password": "admin123"
}
```

#### Success Response:
- **Status:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2026-10-11T14:30:00Z"
}
```

#### Error Responses:
- **400 Bad Request:** Malformed JSON body (`{"error": "invalid request body"}`)
- **401 Unauthorized:** Invalid email or password (`{"error": "invalid credentials"}`)

---

### 4.3. List Projects
Fetches all projects ordered by creation date descending.

- **Method:** `GET`
- **Endpoint:** `/api/projects`
- **Auth Required:** No
- **Headers:** `Accept: application/json`

#### Success Response:
- **Status:** `200 OK`
```json
[
  {
    "id": "ledger-api-mn2a",
    "title": "Ledger API",
    "tagline": "High-throughput financial ledger",
    "year": "2024",
    "category": "systems",
    "status": "completed",
    "featured": true,
    "image": "https://res.cloudinary.com/dfdzkkjqm/image/upload/v1791638927/portfolio/sample.jpg",
    "imageAlt": "Ledger API Dashboard",
    "description": "Designed and implemented an immutable distributed ledger with sub-millisecond query performance.",
    "role": "Lead Backend Engineer",
    "stack": ["Go", "PostgreSQL", "Docker"],
    "metrics": [
      {
        "value": "100k",
        "label": "Ops per second"
      }
    ],
    "liveUrl": "https://example.com",
    "codeUrl": "https://github.com/example/ledger"
  }
]
```

---

### 4.4. Get Project by ID
Retrieves a single project by its unique slug/identifier.

- **Method:** `GET`
- **Endpoint:** `/api/projects/:id`
- **Auth Required:** No
- **URL Parameters:**
  - `id` (`string`, required): The project ID (e.g. `ledger-api-mn2a`).

#### Success Response:
- **Status:** `200 OK`
```json
{
  "id": "ledger-api-mn2a",
  "title": "Ledger API",
  "tagline": "High-throughput financial ledger",
  "year": "2024",
  "category": "systems",
  "status": "completed",
  "featured": true,
  "image": "https://res.cloudinary.com/dfdzkkjqm/image/upload/v1791638927/portfolio/sample.jpg",
  "imageAlt": "Ledger API Dashboard",
  "description": "Designed and implemented an immutable distributed ledger with sub-millisecond query performance.",
  "role": "Lead Backend Engineer",
  "stack": ["Go", "PostgreSQL", "Docker"],
  "metrics": [
    {
      "value": "100k",
      "label": "Ops per second"
    }
  ],
  "liveUrl": "https://example.com",
  "codeUrl": "https://github.com/example/ledger"
}
```

#### Error Response:
- **404 Not Found:** `{"error": "project not found"}`

---

### 4.5. Create Project
Creates a new project record. The server automatically generates a unique `id` based on the title and timestamp.

- **Method:** `POST`
- **Endpoint:** `/api/admin/projects`
- **Auth Required:** Yes (`Authorization: Bearer <token>`)
- **Headers:** `Content-Type: application/json`

#### Request Payload:
| Field | Type | Required | Constraints & Validation |
| :--- | :--- | :--- | :--- |
| `title` | `string` | Yes | Cannot be empty |
| `tagline` | `string` | Yes | Cannot be empty |
| `year` | `string` | Yes | Must be a 4-digit year (regex: `^\d{4}$`) |
| `category` | `string` | Yes | Must be one of: `"creative"`, `"functional"`, `"systems"` |
| `status` | `string` | Yes | Must be one of: `"completed"`, `"ongoing"`, `"ideating"`, `"paused"` |
| `featured` | `boolean` | No | Defaults to `false` |
| `image` | `string` | No | If provided, must be a valid `http://` or `https://` URL |
| `imageAlt` | `string` | No | Descriptive text for the image |
| `description`| `string` | Yes | Minimum 20 characters length |
| `role` | `string` | No | Engineering / design role |
| `stack` | `string[]` | No | Array of technology tags (duplicates are deduplicated) |
| `metrics` | `Metric[]` | No | Array of `{ "value": string, "label": string }` (maximum 3 items) |
| `liveUrl` | `string` | No | If provided, must be a valid URL starting with `http://` or `https://` |
| `codeUrl` | `string` | No | If provided, must be a valid URL starting with `http://` or `https://` |

```json
{
  "title": "Interactive 3D Visualizer",
  "tagline": "Real-time GPU accelerated audio visualizer",
  "year": "2024",
  "category": "creative",
  "status": "completed",
  "featured": true,
  "image": "https://res.cloudinary.com/dfdzkkjqm/image/upload/v1791638927/portfolio/visualizer.webp",
  "imageAlt": "3D Visualizer preview",
  "description": "Exploration of procedural shaders, WebAudio API, and real-time reactive geometry rendering.",
  "role": "Creative Developer",
  "stack": ["Three.js", "WebAudio", "GLSL"],
  "metrics": [
    {
      "value": "60fps",
      "label": "Stable render performance"
    }
  ],
  "liveUrl": "https://visualizer.example.com",
  "codeUrl": "https://github.com/example/visualizer"
}
```

#### Success Response:
- **Status:** `201 Created`
- Returns the created `Project` object including its generated `id`.

#### Error Responses:
- **401 Unauthorized:** Missing or invalid Bearer token
- **400 Bad Request:** Malformed JSON syntax
- **422 Unprocessable Entity:** Payload validation error (returns `{"errors": { ... }}`)

---

### 4.6. Update Project
Updates all fields of an existing project by its ID.

- **Method:** `PUT`
- **Endpoint:** `/api/admin/projects/:id`
- **Auth Required:** Yes (`Authorization: Bearer <token>`)
- **Headers:** `Content-Type: application/json`
- **URL Parameters:**
  - `id` (`string`, required): Project ID to update.

#### Request Payload:
Same fields and validation rules as **Create Project** (see Section 4.5).

#### Success Response:
- **Status:** `200 OK`
- Returns the updated `Project` object.

#### Error Responses:
- **401 Unauthorized:** Missing or invalid Bearer token
- **404 Not Found:** `{"error": "project not found"}`
- **422 Unprocessable Entity:** Validation errors

---

### 4.7. Delete Project
Permanently deletes a project.

- **Method:** `DELETE`
- **Endpoint:** `/api/admin/projects/:id`
- **Auth Required:** Yes (`Authorization: Bearer <token>`)
- **URL Parameters:**
  - `id` (`string`, required): Project ID to delete.

#### Success Response:
- **Status:** `204 No Content` (Empty response body)

#### Error Responses:
- **401 Unauthorized:** Missing or invalid Bearer token
- **404 Not Found:** `{"error": "project not found"}`

---

### 4.8. Upload Image (Cloudinary)
Uploads an image asset to Cloudinary (under the configured folder `portfolio`) and returns the permanent CDN URL for use in `project.image`.

- **Method:** `POST`
- **Endpoint:** `/api/admin/uploads`
- **Auth Required:** Yes (`Authorization: Bearer <token>`)
- **Content-Type:** `multipart/form-data`

#### Request Payload (FormData):
| Key | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `file` | `Binary File` | Yes | Image file (JPEG, PNG, WebP, GIF) |

- **Max File Size:** `5 MB`
- **Accepted MIME Types:** `image/jpeg`, `image/png`, `image/webp`, `image/gif` *(SVG is rejected for script safety)*

#### Success Response:
- **Status:** `201 Created`
```json
{
  "url": "https://res.cloudinary.com/dfdzkkjqm/image/upload/v1791638927/portfolio/jfhp1hz2vetelfmbqz7i.gif"
}
```

#### Error Responses:
- **400 Bad Request:** Missing form field `file` (`{"error": "attach an image as form field \"file\""}`)
- **401 Unauthorized:** Missing or invalid Bearer token
- **413 Request Entity Too Large:** File exceeds 5MB (`{"error": "file is too large"}`)
- **415 Unsupported Media Type:** Invalid image format (`{"error": "only JPEG, PNG, WebP or GIF images are allowed"}`)

---

## 5. TypeScript Contract Types

Frontend applications can import or define these exact types:

```typescript
export interface Metric {
  value: string;
  label: string;
}

export type ProjectCategory = 'creative' | 'functional' | 'systems';
export type ProjectStatus = 'completed' | 'ongoing' | 'ideating' | 'paused';

export interface Project {
  id: string;
  title: string;
  tagline: string;
  year: string;
  category: ProjectCategory;
  status: ProjectStatus;
  featured?: boolean;
  image?: string;
  imageAlt?: string;
  description: string;
  role: string;
  stack: string[];
  metrics: Metric[];
  liveUrl?: string;
  codeUrl?: string;
}

export interface ProjectInput {
  title: string;
  tagline: string;
  year: string;
  category: ProjectCategory;
  status: ProjectStatus;
  featured?: boolean;
  image?: string;
  imageAlt?: string;
  description: string;
  role: string;
  stack: string[];
  metrics: Metric[];
  liveUrl?: string;
  codeUrl?: string;
}

export interface LoginResponse {
  token: string;
  expiresAt: string;
}

export interface UploadResponse {
  url: string;
}

export interface ApiError {
  error: string;
}

export interface ApiValidationError {
  errors: Record<string, string>;
}
```
