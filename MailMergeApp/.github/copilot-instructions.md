**Project Summary**: MailMerge Go is a Windows desktop app (Wails) combining a Go backend (COM automation for Outlook) and a React frontend. Backend lives under `backend/` and exposes Wails methods via `App` bindings in `app.go`.

**Quick Start Commands**:
- **Local dev**: `wails dev` (runs Vite dev server + Go service - hot reload)
- **Build production**: `wails build`
- **Frontend only**: `cd frontend && npm install && npm run dev` or `npm run build`

**Architecture Notes**:
- **Backend**: `MailMergeApp/app.go` binds Go methods to the frontend; core services: `backend/services/{file_service.go, merge_service.go, outlook_service.go}`; models in `backend/models/contact.go`.
- **Frontend**: React + Vite in `frontend/`; generated Wails bindings in `frontend/wailsjs/go/*`; the UI interacts with Go via functions exported in `App` (example: `SelectContactFile`, `ParseContactFile`, `SendBulkEmails`).
- **Event system**: Backend uses `runtime.EventsEmit(ctx, "email:progress", update)`; frontend subscribes with `EventsOn('email:progress', cb)`; payload uses `models.ProgressUpdate`.

**API / Wails Conventions**:
- Any exported method (capitalized) on the `App` struct in `app.go` becomes callable from the frontend via `wailsjs`.
- Keep APIs typed using `backend/models` structs and use the generated bindings under `frontend/wailsjs/go/models` for requests/responses.
- For new APIs: 1) add an exported method to `App` with a Request/Response model under `backend/models`; 2) add the service implementation under `backend/services/`; 3) call `wails build` (or `wails dev`) to regenerate `wailsjs` bindings.

**Important Implementation Patterns**:
- **COM/Outlook**: `backend/services/outlook_service.go` uses `github.com/go-ole/go-ole` and must initialize COM per thread with `ole.CoInitializeEx`. Outlook calls are serialized with a mutex `s.mu` — follow this pattern and avoid parallelizing Outlook operations.
- **Event emission**: Use `runtime.EventsEmit(ctx, "email:progress", update)` to stream progress to the UI. Match fields in `backend/models.ProgressUpdate`.
- **File parsing**: `FileService` accepts `.csv` and `.xlsx` and detects `email`, `firstname`, and `lastname` case-insensitively — replicate the column normalizing logic if adding more importers.
- **Template parsing**: `MergeService` uses regex `\{(\w+)\}` and supports only `{FirstName}`, `{LastName}`, `{Email}`. Use `ValidateTemplate` to report unknown merge fields.

**Safety & Testing Tips**:
- Outlook COM runs only on Windows; automated tests against `OutlookService` must mock COM interactions. The service exposes `osStat` and `osIsNotExist` variables for easy attachment testing.
- Protect COM usage: initialize and uninitialize per goroutine and keep operations in a serialized block (see `SendBulkEmails`).
- Validate attachments and file paths before attempting to use them (see `OutlookService.validateAttachment`).

**Developer Workflow**:
- Add a new backend service: `backend/services/your_service.go` and export it via `App` in `app.go` by adding a field and initializing in `NewApp()`.
- Add TypeScript typings: After `wails build` or `wails dev`, generated bindings in `frontend/wailsjs` will reflect new types/methods.
- Frontend: import methods like `import { SendBulkEmails } from '../wailsjs/go/main/App';` and call with typed models `wailsjs/go/models`.

**Code Style & Conventions**:
- Keep logic for device/OS-specific things contained in services (e.g., COM, file i/o). Avoid leaking platform checks into shared logic.
- All public app actions should be on `App` (app.go) and named clearly, e.g., `SelectContactFile`, `ParseContactFile`, `SendBulkEmails`.
- Follow the `models` package for serializable request/response shapes used by the frontend.

**Where to Look for More Context**:
- App & server entrypoints: [main.go](main.go), [app.go](app.go)
- Core services and models: [backend/services](backend/services), [backend/models/contact.go](backend/models/contact.go)
- Frontend integration examples: [frontend/src/App.tsx](frontend/src/App.tsx), [frontend/src/components](frontend/src/components)
- Example inputs: [samples/contacts.csv](samples/contacts.csv)
- Wails config: [wails.json](wails.json)

If anything is unclear or you want more examples (e.g., adding a new API or mocking COM in tests), tell me which part to expand and I’ll add focused examples and snippets.
