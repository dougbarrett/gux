# Feature Landscape: File Upload System for Gux

**Domain:** File upload system for a full-stack Go/WASM framework with admin scaffolding
**Researched:** 2026-02-01
**Overall Confidence:** HIGH (patterns verified across Django, Laravel, Rails, Payload CMS, and Go ecosystem)

---

## Table Stakes

Features users expect from any file upload system in a web framework. Missing any of these means the feature feels incomplete or unusable.

| Feature | Why Expected | Complexity | Dependencies | Notes |
|---------|--------------|------------|--------------|-------|
| **Click-to-browse file selection** | Most basic upload interaction; HTML `<input type="file">` | Low | None | Foundation for everything else |
| **Drag-and-drop zone** | Every modern upload UI supports drag-and-drop; users assume it works | Medium | `ui.FileUpload` component | Use HTML5 Drag and Drop API; visual feedback on dragover is critical |
| **Upload progress indicator** | Users need to know something is happening, especially for large files | Medium | WASM XMLHttpRequest or ReadableStream | Standard `fetch` does not emit upload progress in all browsers; may need XMLHttpRequest fallback |
| **File type validation (client-side)** | Prevent wrong file types before upload starts; avoids wasted bandwidth | Low | `ui.FileUpload` component | Accept attribute + JS validation; validate MIME type, not just extension |
| **File size validation (client-side)** | Prevent oversized files before upload; clear error messaging | Low | `ui.FileUpload` component | Check `File.size` before uploading; configurable max size |
| **Server-side file type validation** | Client validation is bypassable; server must re-validate | Low | Upload endpoint | Check MIME type via magic bytes (`http.DetectContentType` in Go), not just Content-Type header |
| **Server-side file size limits** | Protect against abuse; configurable per-field or global | Low | Upload endpoint, `core.App` config | Use `http.MaxBytesReader` in Go |
| **Single file upload per field** | Most file fields are single-file (avatar, document) | Low | `"type": "file"` in config | Store as string path in DB |
| **Multiple file upload per field** | Photo galleries, document collections | Medium | `"type": "file[]"` in config | Store as JSON array of paths; need add/remove UX |
| **Image preview before upload** | Users expect to see what they selected before committing | Medium | `ui.FileUpload` component | Use `URL.createObjectURL()` in WASM; more efficient than `FileReader.readAsDataURL()` |
| **Storage abstraction (local + S3)** | Local for dev, S3 for prod is the universal pattern (Django, Laravel, Rails all do this) | Medium | New `storage` package | Interface: `Put(key, reader) -> URL`, `Delete(key)`, `URL(key)` |
| **File serving endpoint** | Uploaded files need to be accessible via URL | Low | `core.App` routing | Local: serve from disk; S3: redirect to signed URL or serve public URL |
| **CRUD integration in admin forms** | File fields must appear in generated create/edit forms automatically | High | `gux gen`, `ui.FileUpload`, admin page templates | Core value prop: `"type": "file"` in gux.config.json triggers file upload widget |
| **File display in admin detail views** | Show filename/link for documents, thumbnail for images | Medium | Admin detail page templates | Detect image vs non-image by stored MIME type or extension |
| **File display in admin list views** | Show thumbnail or file icon in table rows | Medium | Admin list page templates, DTO generation | Small thumbnails or file type icons in table cells |
| **Delete/replace existing file** | Users must be able to remove or replace an uploaded file | Medium | Upload endpoint, storage abstraction | Must clean up old file from storage when replacing; show current file with remove button |
| **Unique filename generation** | Prevent collisions and path traversal attacks | Low | Upload endpoint | Use UUID-based naming; preserve original filename in metadata |
| **Configurable upload path** | Different models store files in different paths | Low | gux.config.json field options | e.g., `"path": "avatars"` stores in `uploads/avatars/` |
| **CSRF protection on upload** | Upload endpoints are POST; CSRF is mandatory | Low | Already exists in Gux | Gux CSRF is automatic; upload endpoint just needs to be a standard POST route |

---

## Differentiators

Features that set Gux apart from other frameworks. Not expected but add significant value.

| Feature | Value Proposition | Complexity | Dependencies | Notes |
|---------|-------------------|------------|--------------|-------|
| **Zero-config file fields via gux.config.json** | Add `"type": "file"` to any field and get complete upload UI, storage, API, and admin integration with zero code. Django's `ImageField` equivalent but fully declarative. | High | All table stakes features | This is the killer feature. Other Go frameworks require manual wiring of upload handlers, storage, forms, and display. |
| **Image thumbnail generation on upload** | Auto-generate smaller versions for list views and previews; reduces bandwidth | Medium | Image processing library (`disintegration/imaging` or `nfnt/resize`) | Generate 2 sizes: thumbnail (150px) and medium (600px); store alongside original |
| **Presigned URL uploads (direct-to-S3)** | Client uploads directly to S3, bypassing server bandwidth; critical for large files | High | S3 storage driver, presigned URL generation endpoint | Server generates presigned POST; client uploads directly; server records metadata after confirmation. Major performance win for production. |
| **Lifecycle hooks (BeforeUpload, AfterUpload, BeforeDelete)** | Custom validation, processing, or side-effects without modifying generated code | Medium | Hook registration (matches existing `WithCreateHook` pattern) | e.g., `WithBeforeUpload(func(file FileInfo) error)`, `WithAfterUpload(func(file StoredFile) error)` |
| **Automatic orphan cleanup on record delete** | Delete files from storage when parent record is deleted via CRUD | Medium | CRUD delete hooks, storage abstraction | Hook into CRUD delete; configurable (some want soft-delete with file retention) |
| **File metadata storage** | Store original filename, MIME type, file size, and image dimensions alongside the path | Low-Medium | Migration generation, DTO field updates | Two approaches: (a) dedicated `gux_files` table with polymorphic relation (Rails Active Storage pattern) or (b) JSON metadata column. Recommend approach (a) for queryability. |
| **Image cropping in admin UI** | Let admins crop avatars or hero images before upload | High | JS cropping library (Cropper.js) + `core.LoadScript` | Significant JS integration complexity in WASM; good candidate for a hook-based approach where users opt in |
| **Drag-and-drop reordering for multi-file** | Let users reorder images in a gallery field | Medium | `ui.FileUpload` component, sortable interaction | Store order as array position; useful for product images, portfolios |
| **Storage config in gux.config.json** | Define storage backend in config, not just code | Low | Config parser update | e.g., `"storage": {"driver": "s3", "bucket": "my-app", "region": "us-east-1"}` with env var interpolation |
| **Accept filter presets** | Shorthand `"accept": "images"` expands to common image MIME types; `"accept": "documents"` for PDF/DOC/etc. | Low | Config parser, `ui.FileUpload` | Reduces config boilerplate; still allow explicit MIME type lists |

---

## Anti-Features

Features to deliberately NOT build. Common mistakes in this domain that add complexity without proportional value.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Built-in media library / asset manager** | Massive scope; becomes its own product (see Payload CMS, WordPress Media Library). A framework should not ship a content management system. | Provide file fields on models. If users need a media library, they build one using CRUD + file fields. |
| **Built-in image transformation CDN (resize on request)** | Complex infrastructure concern; better handled by Imgix, Cloudinary, or CloudFront functions. Server-side on-demand resizing burns CPU and invites abuse. | Generate fixed thumbnail sizes at upload time. Provide `AfterUpload` hook for users to add custom transformation pipelines. |
| **Virus/malware scanning** | Infrastructure concern, not framework concern. Requires external services (ClamAV, AWS Macie). Adding it creates a false sense of security. | Document that users should add scanning in `BeforeUpload` hook or use S3 event triggers with Lambda/similar. |
| **In-browser video transcoding** | Extremely complex, unreliable in WASM, burns user CPU and battery. | Accept video files as-is. Recommend external transcoding services (AWS MediaConvert, FFmpeg workers) via `AfterUpload` hook. |
| **Universal inline file preview (PDF, Office docs, etc.)** | Previewing PDFs and Office documents requires heavy JS libraries (pdf.js, Office viewer embeds) or external services. Scope creep. | Preview images only (native `<img>` tag). Show filename + download link + file type icon for everything else. |
| **Per-file ACL / permission system** | Files should inherit their parent model's access control. Adding per-file permissions creates a parallel permission system that nobody wants to maintain. | Files inherit CRUD model auth (public/roles). If users need per-file ACLs, they build a custom model for it. |
| **Chunked/resumable uploads (TUS protocol) in v1** | Only needed for very large files (100MB+). Adds protocol complexity, requires server-side state management, and most admin uploads are under 50MB. | Support standard multipart upload. Add resumable uploads as a future enhancement when users request it for specific use cases (video platforms, large dataset uploads). |
| **Real-time collaborative file editing** | Google Docs territory. Completely out of scope for an upload system. | Provide download link. |
| **Automatic CDN cache invalidation** | CDN configuration is infrastructure, not framework business. Too many CDN providers with different APIs. | Use content-hash-based filenames (e.g., `abc123_avatar.jpg`) for automatic cache busting. Document CDN setup in guides. |

---

## Feature Dependencies

```
gux.config.json "type": "file" / "type": "file[]"
    |
    v
gux gen (code generation)
    |
    +---> Model field (string path in DB)
    +---> DTO field (URL string or file metadata struct)
    +---> Upload API endpoint (POST /__gux_api/upload/:model/:field)
    +---> Admin form integration (ui.FileUpload component)
    +---> Admin detail/list integration (image preview or file link)
    |
    v
Storage Abstraction Layer (storage package)
    |
    +---> Storage interface: Put / Delete / URL
    +---> Local driver (filesystem, dev/simple deploys)
    +---> S3 driver (production, scalable)
    +---> Config: gux.config.json + env vars + code overrides
    |
    v
ui.FileUpload Component
    |
    +---> Click-to-browse (basic HTML input, no JS deps)
    +---> Drag-and-drop zone (HTML5 DnD API in WASM)
    +---> Client-side validation (type + size, before upload)
    +---> Progress indicator (XMLHttpRequest or ReadableStream)
    +---> Image preview (URL.createObjectURL in WASM)
    +---> Current file display + remove button (edit mode)
    +---> Multi-file mode (for "file[]" type)
    |
    v
Lifecycle Hooks (optional, user-configured)
    +---> BeforeUpload: validate, reject, transform
    +---> AfterUpload: post-process, generate thumbnails, notify
    +---> BeforeDelete: archive, cleanup references
    |
    v
Thumbnail Generation (optional, images only)
    +---> On upload: generate thumbnail (150px) + medium (600px)
    +---> Store alongside original in same storage backend
    +---> Serve in admin list/detail views
```

**Critical path:** Storage abstraction and `ui.FileUpload` component are independent workstreams that can be built in parallel. They converge at the upload API endpoint, which then feeds into code generation (`gux gen`).

**Dependency on existing Gux features:**
- CSRF protection: Already automatic; upload endpoint inherits it
- CRUD model registration: File upload hooks attach to existing `CRUDModel` struct
- Admin page hooks: File display in admin uses existing hook slot pattern
- `core.LoadScript`: Already exists for dynamic JS loading (useful for optional Cropper.js)
- `ui.FormField`: FileUpload wraps in existing FormField for consistent label/error display
- `ui.Input` pattern: FileUpload follows same Props struct + Tailwind + SSR/WASM pattern

---

## MVP Recommendation

For MVP, prioritize in this order:

### Must Ship (MVP)

1. **Storage abstraction with Local + S3 drivers** -- Foundation everything else builds on. Interface: `Put(ctx, key, reader, opts) (string, error)`, `Delete(ctx, key) error`, `URL(key) string`. Get the interface right because changing it later means rewriting all integrations.

2. **ui.FileUpload component (single file)** -- Click-to-browse, drag-and-drop, client-side validation (type + size), progress bar, image preview, current file display with remove. Follow existing `ui.Input` Props pattern. Must work in both SSR (render static upload zone) and WASM (interactive).

3. **Upload API endpoint** -- `POST /__gux_api/upload/:model/:field` that validates, stores via storage abstraction, and returns the file URL/path. Follows existing CRUD endpoint patterns. CSRF is automatic.

4. **gux.config.json field types** -- `"type": "file"` (single) and `"type": "file[]"` (multiple). Config options: `"accept"`, `"maxSize"`, `"path"`. `gux gen` produces correct model fields, DTO fields, upload endpoints, form widgets, and display code.

5. **Admin form/detail/list integration** -- Generated admin pages use `ui.FileUpload` in create/edit forms, show image thumbnails or file links in detail views, and display small thumbnails or file icons in list table cells.

6. **Lifecycle hooks** -- `WithBeforeUpload(func)`, `WithAfterUpload(func)`, `WithBeforeDelete(func)`. Follow existing `WithCreateHook` pattern. Critical: ensure BeforeDelete runs BEFORE physical file deletion (learn from Strapi's bug where file was deleted before hook ran).

### Defer to Post-MVP

- **Presigned URL (direct-to-S3) uploads**: Significant complexity; standard proxy upload handles files under 100MB fine, which covers 95%+ of admin use cases
- **Image thumbnail generation**: Can ship without it if images display at constrained CSS sizes; add server-side generation as fast follow using `disintegration/imaging`
- **Image cropping UI**: Nice-to-have; users can add via `AfterUpload` hook + `core.LoadScript` for Cropper.js
- **Drag-and-drop reordering for multi-file**: Enhancement to `ui.FileUpload` component; not blocking for initial multi-file support
- **File metadata table**: Start with string path in model field; upgrade to metadata table if users need to query by file type/size
- **Chunked/resumable uploads**: Only when users report need for large file upload scenarios

---

## Sources

### Upload UI Patterns
- [Salt Design System - File Upload Pattern](https://www.saltdesignsystem.com/salt/patterns/file-upload) - Upload area + file list pattern
- [PatternFly - Multiple File Upload](https://www.patternfly.org/components/file-upload/multiple-file-upload/design-guidelines/) - Multi-file upload states and feedback
- [web.dev - Drag and Drop Files](https://web.dev/patterns/files/drag-and-drop-files) - HTML5 DnD API reference
- [Fine Uploader - Thumbnails](https://docs.fineuploader.com/features/thumbnails.html) - Client-side preview generation patterns

### Storage and Backend Patterns
- [Payload CMS - Upload Documentation](https://payloadcms.com/docs/upload/overview) - Media collection pattern, hook-based remote storage
- [QuickAdminPanel - File/Photo Upload Fields](https://helpdocs.quickadminpanel.com/create-panel/file-photo-upload-fields) - Admin CRUD file field integration
- [Laravel Admin - Model Form Upload](https://laravel-admin.org/docs/en/model-form-upload) - Framework-level file field configuration
- [Django ImageField](https://nemecek.be/blog/79/working-with-django-imagefield) - Model field type pattern
- [Gin + S3/Local Storage Tutorial](https://www.djamware.com/post/684ce060e5e3c12b1f713f06/build-a-file-upload-api-in-go-with-gin-framework-and-s3-or-local-storage) - Go storage backend switching

### Lifecycle and Processing
- [Strapi Upload Lifecycle Issue #17170](https://github.com/strapi/strapi/issues/17170) - Critical ordering bug: file deleted before BeforeDelete hook runs
- [Supabase Storage - Upload Processing](https://deepwiki.com/supabase/storage/3.3-upload-processing) - TUS lifecycle hooks, advisory locks
- [AWS S3 Best Practices 2025](https://howik.com/best-practices-for-aws-s3-storage) - Multipart uploads, lifecycle rules

### Architecture
- [Three Dots Labs - Go Web Anti-Patterns](https://threedots.tech/post/common-anti-patterns-in-go-web-applications/) - Separation of concerns, avoiding model coupling
