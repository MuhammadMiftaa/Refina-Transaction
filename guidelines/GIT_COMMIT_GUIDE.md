# Git Commit Message Guide

Dokumen ini adalah panduan lengkap penulisan commit message untuk project ini. Ikuti aturan ini secara konsisten agar riwayat git mudah dibaca, di-search, dan di-audit.

---

## Format Dasar

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

### Contoh

```
feat(wallet): add CreateWalletGRPC method with initial deposit

fix(auth): correct JWT expiry validation logic

refactor(service): extract outbox publishing into separate method

test(wallet): add unit tests for service layer with 91.5% coverage

chore: update go.mod to Go 1.25.0
```

---

## Type (Wajib)

Pilih **tepat satu** dari daftar berikut. Tidak ada tipe lain.

| Type       | Kapan Digunakan                                                       | Contoh                                                 |
| ---------- | --------------------------------------------------------------------- | ------------------------------------------------------ |
| `feat`     | Menambahkan fitur baru atau behavior baru                             | Menambah endpoint, method baru, integrasi baru         |
| `fix`      | Memperbaiki bug atau kesalahan logic                                  | Fix nil pointer, salah kondisi, salah port             |
| `refactor` | Mengubah kode tanpa mengubah behavior (tidak feat, bukan fix)         | Ekstrak method, rename variable, restructure           |
| `test`     | Menambah atau mengubah test — **tidak ada perubahan production code** | Buat unit test, fix failing test, tambah coverage      |
| `chore`    | Tugas maintenance yang tidak mempengaruhi kode aplikasi               | Update dependency, update `.gitignore`, konfigurasi CI |
| `docs`     | Perubahan dokumentasi saja                                            | Update README, tambah komentar, buat guide             |
| `perf`     | Perubahan yang meningkatkan performa                                  | Optimasi query, cache, algoritma                       |
| `style`    | Perubahan formatting/whitespace — tanpa mengubah logic                | Gofmt, perapian indentasi                              |
| `ci`       | Perubahan pada konfigurasi CI/CD                                      | Update Dockerfile, workflow GitHub Actions             |
| `revert`   | Revert commit sebelumnya                                              | `revert: feat(wallet): add CreateWalletGRPC`           |

---

## Scope (Opsional, Sangat Direkomendasikan)

Scope menjelaskan **bagian mana dari codebase** yang diubah. Tulis dalam tanda kurung setelah type.

### Scope yang Valid di Project Ini

| Scope         | Komponen                                             |
| ------------- | ---------------------------------------------------- |
| `wallet`      | Service, repository, handler, test untuk wallet      |
| `wallet-type` | Service, repository, handler, test untuk wallet type |
| `outbox`      | Outbox publisher, repository                         |
| `grpc`        | gRPC server atau client                              |
| `http`        | HTTP handler, middleware, router                     |
| `queue`       | RabbitMQ client                                      |
| `db`          | Database config, migration, seeder                   |
| `auth`        | Autentikasi, middleware, interceptor                 |
| `config`      | Konfigurasi env, logger                              |
| `log`         | Logger setup dan utility                             |

### Contoh dengan Scope

```
feat(wallet): implement DeleteWallet with balance validation
fix(grpc): correct nil pointer in transaction interceptor
test(outbox): add context cancellation test for publisher
refactor(wallet-type): extract validation logic into helper
chore(db): add migration for outbox_messages table
```

---

## Subject (Wajib)

Aturan penulisan subject line:

```
✅ feat(wallet): add unit tests for service layer
✅ fix(grpc): correct nil pointer dereference in interceptor
✅ refactor(outbox): extract publishMessage into separate method

❌ feat(wallet): Added some tests          (jangan past tense)
❌ feat(wallet): Add tests.                (jangan pakai titik)
❌ feat: wallet                            (terlalu pendek, tidak deskriptif)
❌ feat(wallet): add tests for the wallet service layer in Go (terlalu panjang)
```

| Aturan                   | Detail                                                     |
| ------------------------ | ---------------------------------------------------------- |
| **Imperative mood**      | "add", "fix", "update", "remove" — bukan "added", "fixed"  |
| **Huruf kecil**          | Mulai dengan huruf kecil setelah tanda titik dua dan spasi |
| **Tanpa titik** di akhir | Tidak ada tanda baca di akhir subject line                 |
| **Maks 72 karakter**     | Jika lebih, pindahkan detail ke body                       |
| **Bahasa Inggris**       | Selalu gunakan bahasa Inggris                              |

---

## Body (Opsional)

Gunakan body jika perubahan membutuhkan konteks tambahan. Pisahkan dari subject dengan **satu baris kosong**.

```
refactor(wallet): extract outbox message creation into helper

Previously, the outbox message creation logic was duplicated
across CreateWallet, UpdateWallet, and DeleteWallet methods.
Extracted into buildOutboxMessage() to reduce duplication.

No behavior change.
```

Gunakan body untuk menjelaskan:

- **Kenapa** perubahan ini dibuat (bukan apa yang diubah)
- Trade-off atau keputusan desain yang penting
- Konteks yang tidak terlihat dari kode

---

## Footer (Opsional)

Gunakan footer untuk referensi issue atau breaking change.

```
feat(auth): add JWT refresh token support

Closes #42
BREAKING CHANGE: /auth/login now returns refresh_token field
```

---

## Aturan Tambahan

### 1. Satu Perubahan, Satu Commit

```bash
# BURUK — terlalu banyak hal dalam satu commit
feat(wallet): add CRUD, fix bug in auth, update README

# BAGUS — satu fokus per commit
feat(wallet): add CreateWallet with transaction support
fix(auth): handle expired token in gRPC interceptor
docs: update README with gRPC setup instructions
```

### 2. Jangan Commit File yang Tidak Relevan

File yang **tidak boleh masuk** commit:

- `.env` (sudah ada di `.gitignore`)
- `coverage.out`, `coverage.html`
- File editor seperti `.vscode/settings.json` (kecuali memang project-level config)
- Binary atau build artifact

### 3. Test Tidak Boleh Digabung dengan Feature

```bash
# BURUK
feat(wallet): add CreateWallet and its unit tests

# BAGUS — pisahkan
feat(wallet): add CreateWallet with transaction support
test(wallet): add unit tests for CreateWallet service method
```

Pengecualian: test kecil (1-2 fungsi) untuk validasi langsung boleh digabung jika perubahan sangat kecil.

### 4. Referensi Jika Ada Perubahan Breaking

```
feat(wallet): change CreateWallet to require non-zero balance

BREAKING CHANGE: CreateWallet now returns error if balance == 0.
Previously allowed zero balance on creation.
```

---

## Contoh Nyata dari Project Ini

Commit message yang **baik** (sudah ada di history):

```
feat: remove unused logging import from user_metadata.go
feat: refactor wallet service to remove repository dependency and enhance wallet response structure
fix: correct HTTP server setup condition to ensure proper initialization
fix: correct gRPC server port binding to include TCP protocol
refactor: update gRPC server setup to handle errors and modify wallet service methods
chore: clean up code structure and remove unused code blocks
```

Commit message yang **perlu diperbaiki** (ada di history):

```
# Terlalu generik tanpa scope
upgrade protobuf
→ seharusnya: chore(deps): upgrade Refina-Protobuf to v1.7.1

# Tidak menggunakan conventional format
Refactor main logging
→ seharusnya: refactor(log): consolidate logger setup in main

# Tidak menggunakan conventional format
Implement structured logging across wallet and wallet type services
→ seharusnya: feat(log): add structured logging to wallet and wallet-type services
```

---

## Quick Reference

```bash
# Feature baru
git commit -m "feat(wallet): add GetWalletsByUserIDGroupByType endpoint"

# Bug fix
git commit -m "fix(outbox): prevent nil panic when channel is unavailable"

# Refactor
git commit -m "refactor(service): extract common outbox message builder"

# Unit test
git commit -m "test(wallet): add unit tests for service layer with 91.5% coverage"

# Dependency update
git commit -m "chore(deps): upgrade testify to v1.11.1"

# Migration
git commit -m "chore(db): add migration for outbox_messages table"

# Dokumentasi
git commit -m "docs: add unit test guide and git commit message guide"

# CI/CD
git commit -m "ci: add Go test step to GitHub Actions workflow"
```

---

## Checklist Sebelum Commit

- [ ] Type sudah dipilih dari daftar yang valid
- [ ] Subject menggunakan imperative mood ("add", bukan "added")
- [ ] Subject tidak lebih dari 72 karakter
- [ ] Subject tidak diawali huruf kapital (kecuali proper noun)
- [ ] Subject tidak diakhiri tanda titik
- [ ] Scope mencerminkan komponen yang benar-benar diubah
- [ ] Tidak ada file yang tidak relevan ikut ter-staging
- [ ] Satu commit fokus pada satu perubahan logis
