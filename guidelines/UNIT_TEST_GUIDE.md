# Unit Test Guide — Go Service Layer

Dokumen ini adalah panduan lengkap dan opinionated untuk menulis unit test pada **service layer** di project Go dengan arsitektur Repository Pattern. Semua aturan diambil langsung dari implementasi nyata di project `refina-wallet`.

---

## Daftar Isi

1. [Prinsip Utama](#1-prinsip-utama)
2. [Dependencies & Setup](#2-dependencies--setup)
3. [Struktur Direktori](#3-struktur-direktori)
4. [Cara Membuat Mock](#4-cara-membuat-mock)
5. [Pola Test File](#5-pola-test-file)
6. [Pola Test Case](#6-pola-test-case)
7. [Checklist Test Case per Method Type](#7-checklist-test-case-per-method-type)
8. [Aturan Penting yang Sering Terlewat](#8-aturan-penting-yang-sering-terlewat)
9. [Menjalankan Test & Cek Coverage](#9-menjalankan-test--cek-coverage)
10. [Referensi Cepat](#10-referensi-cepat)

---

## 1. Prinsip Utama

| #   | Prinsip                                                                             |
| --- | ----------------------------------------------------------------------------------- |
| 1   | **Satu file test per satu file service** — `wallets.go` → `wallets_test.go`         |
| 2   | **Package sama dengan source** — gunakan `package service` (white-box testing)      |
| 3   | **Tidak ada database nyata** — semua dependensi di-mock menggunakan `testify/mock`  |
| 4   | **Setiap method memiliki minimal 2 test** — satu happy path, satu error path        |
| 5   | **Mock harus di-verify** — selalu panggil `AssertExpectations` di akhir setiap test |
| 6   | **Test harus independen** — tidak ada state yang dibagikan antar test case          |
| 7   | **Nama test harus deskriptif** — format `Test<Method>_<Scenario>`                   |
| 8   | **Target coverage ≥ 90%**                                                           |

---

## 2. Dependencies & Setup

### Tambahkan ke `go.mod`

```bash
go get github.com/stretchr/testify
go mod tidy
```

### Package yang digunakan

```go
import (
    "github.com/stretchr/testify/assert" // assertion
    "github.com/stretchr/testify/mock"   // mocking
)
```

### Inisialisasi Logger (jika service menggunakan global logger)

Jika project menggunakan global logger (misalnya `logrus`) yang harus diinisialisasi sebelum digunakan, **wajib** membuat file `service_test.go` berisi `TestMain`:

```go
// internal/service/service_test.go
package service

import (
    "io"
    "os"
    "testing"

    "your-module/config/log"

    "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
    // Inisialisasi logger agar tidak panic saat test
    log.Log = logrus.New()
    log.Log.SetOutput(io.Discard) // buang semua output log saat testing
    log.Log.SetLevel(logrus.PanicLevel)

    os.Exit(m.Run())
}
```

> **Kenapa perlu ini?** Service yang memanggil `log.Error(...)`, `log.Warn(...)`, dll. akan **panic dengan nil pointer dereference** jika logger belum diinisialisasi sebelum test berjalan.

---

## 3. Struktur Direktori

```
internal/
└── service/
    ├── mocks/                          ← semua file mock
    │   ├── mock_wallets_repository.go
    │   ├── mock_wallet_types_repository.go
    │   ├── mock_outbox_repository.go
    │   ├── mock_tx_manager.go          ← termasuk MockTransaction
    │   ├── mock_transaction_client.go  ← mock gRPC / external client
    │   └── mock_rabbitmq_client.go     ← mock message queue
    ├── service_test.go                 ← TestMain (shared setup)
    ├── wallets.go
    ├── wallets_test.go
    ├── walletTypes.go
    ├── walletTypes_test.go
    ├── outboxMessage.go
    └── outboxMessage_test.go
```

**Aturan direktori:**

- Semua mock berada di sub-package `mocks/` di dalam package `service`
- File test menggunakan `package service` (bukan `package service_test`) agar bisa akses internal type
- `service_test.go` hanya berisi `TestMain` — tidak ada test case di sini

---

## 4. Cara Membuat Mock

Setiap interface di layer repository atau external client harus memiliki mock-nya sendiri.

### Template Mock Repository

```go
// internal/service/mocks/mock_<name>_repository.go
package mocks

import (
    "context"

    "your-module/internal/repository"
    "your-module/internal/types/model"

    "github.com/stretchr/testify/mock"
)

type Mock<Name>Repository struct {
    mock.Mock
}

// Implementasikan SEMUA method dari interface
func (m *Mock<Name>Repository) GetAll(ctx context.Context, tx repository.Transaction) ([]model.<Entity>, error) {
    args := m.Called(ctx, tx)
    return args.Get(0).([]model.<Entity>), args.Error(1)
}

func (m *Mock<Name>Repository) GetByID(ctx context.Context, tx repository.Transaction, id string) (model.<Entity>, error) {
    args := m.Called(ctx, tx, id)
    return args.Get(0).(model.<Entity>), args.Error(1)
}

// ... method lainnya
```

### Template Mock TxManager

```go
// internal/service/mocks/mock_tx_manager.go
package mocks

import (
    "context"

    "your-module/internal/repository"

    "github.com/stretchr/testify/mock"
)

// MockTxManager — mock untuk database transaction manager
type MockTxManager struct {
    mock.Mock
}

func (m *MockTxManager) Begin(ctx context.Context) (repository.Transaction, error) {
    args := m.Called(ctx)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(repository.Transaction), args.Error(1)
}

// MockTransaction — mock untuk objek transaksi itu sendiri
type MockTransaction struct {
    mock.Mock
}

func (m *MockTransaction) Commit() error {
    args := m.Called()
    return args.Error(0)
}

func (m *MockTransaction) Rollback() error {
    args := m.Called()
    return args.Error(0)
}
```

> **Perhatian:** `MockTxManager` dan `MockTransaction` berada dalam **satu file** karena keduanya saling berkaitan dan selalu digunakan bersama.

### Template Mock External Client (gRPC / HTTP)

```go
// internal/service/mocks/mock_<name>_client.go
package mocks

import (
    "context"

    pb "github.com/org/proto/service"
    "github.com/stretchr/testify/mock"
)

type Mock<Name>Client struct {
    mock.Mock
}

func (m *Mock<Name>Client) SomeMethod(ctx context.Context, arg string) (*pb.Response, error) {
    args := m.Called(ctx, arg)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*pb.Response), args.Error(1)
}
```

> **Aturan untuk pointer return:** Selalu cek `nil` sebelum type-assert jika return type adalah pointer. Jika tidak, test yang mock return `nil` akan **panic**.

---

## 5. Pola Test File

### Struktur wajib di setiap test file

```go
package service

import (
    "context"
    "errors"
    "testing"
    "time"

    "your-module/internal/service/mocks"
    "your-module/internal/types/dto"
    "your-module/internal/types/model"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// SECTION 1: Test Dependency Container
// ─────────────────────────────────────────────

type <entity>TestDeps struct {
    txManager  *mocks.MockTxManager
    repo       *mocks.Mock<Entity>Repository
    // ... dependency lain
    tx         *mocks.MockTransaction
}

func new<Entity>TestDeps() *<entity>TestDeps {
    return &<entity>TestDeps{
        txManager: new(mocks.MockTxManager),
        repo:      new(mocks.Mock<Entity>Repository),
        tx:        new(mocks.MockTransaction),
    }
}

func (d *<entity>TestDeps) service() <Entity>Service {
    return New<Entity>Service(d.txManager, d.repo /* ... */)
}

func (d *<entity>TestDeps) assertAll(t *testing.T) {
    t.Helper()
    d.txManager.AssertExpectations(t)
    d.repo.AssertExpectations(t)
    d.tx.AssertExpectations(t)
    // ... semua mock lain
}

// ─────────────────────────────────────────────
// SECTION 2: Fixed UUIDs & Timestamps
// ─────────────────────────────────────────────

var (
    entityID  = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
    fixedTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
)

// ─────────────────────────────────────────────
// SECTION 3: Sample Data Factories
// ─────────────────────────────────────────────

func sample<Entity>Model() model.<Entity> {
    return model.<Entity>{
        Base: model.Base{ID: entityID, CreatedAt: fixedTime, UpdatedAt: fixedTime},
        // ... field lain
    }
}

func sample<Entity>Request() dto.<Entity>Request {
    return dto.<Entity>Request{
        // ... field request
    }
}

// ─────────────────────────────────────────────
// SECTION 4: Test Cases (satu section per method)
// ─────────────────────────────────────────────

// =====================================================================
// MethodName
// =====================================================================

func TestMethodName_Success(t *testing.T) { /* ... */ }
func TestMethodName_ErrorScenario(t *testing.T) { /* ... */ }
```

### Kenapa menggunakan `TestDeps` struct?

- **DRY** — tidak perlu `new(mocks.MockX)` berkali-kali
- **Satu tempat** untuk membuat semua mock
- `assertAll()` memastikan **semua** mock diverifikasi, tidak ada yang terlewat
- Mudah menambah dependency baru tanpa refactor setiap test

### Kenapa menggunakan fixed UUIDs?

```go
// BURUK — setiap run menghasilkan nilai berbeda
walletID := uuid.New()

// BAGUS — deterministik, mudah debug
walletID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
```

---

## 6. Pola Test Case

### Happy Path (Success)

```go
func TestGetAllWallets_Success(t *testing.T) {
    // 1. Arrange — setup deps dan mock
    d := newWalletTestDeps()
    svc := d.service()

    wallets := []model.Wallets{sampleWalletModel()}
    d.walletsRepo.On("GetAllWallets", mock.Anything, nil).Return(wallets, nil)

    // 2. Act — panggil method yang ditest
    result, err := svc.GetAllWallets(context.Background())

    // 3. Assert — verifikasi hasil
    assert.NoError(t, err)
    assert.Len(t, result, 1)
    assert.Equal(t, walletID.String(), result[0].ID)
    d.assertAll(t) // selalu di akhir
}
```

### Error Path

```go
func TestGetAllWallets_RepositoryError(t *testing.T) {
    d := newWalletTestDeps()
    svc := d.service()

    d.walletsRepo.On("GetAllWallets", mock.Anything, nil).
        Return([]model.Wallets{}, errors.New("db error"))

    result, err := svc.GetAllWallets(context.Background())

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "get all wallets") // verifikasi error wrapping
    d.assertAll(t)
}
```

### Method dengan Transaction

```go
func TestCreateWallet_Success(t *testing.T) {
    d := newWalletTestDeps()
    svc := d.service()

    // Mock semua langkah dalam urutan yang benar
    d.typesRepo.On("GetWalletTypeByID", mock.Anything, nil, req.WalletTypeID).Return(wt, nil)
    d.txManager.On("Begin", mock.Anything).Return(d.tx, nil)
    d.walletsRepo.On("CreateWallet", mock.Anything, d.tx, mock.Anything).Return(w, nil)
    d.outboxRepo.On("Create", mock.Anything, d.tx, mock.Anything).Return(nil)
    d.tx.On("Commit").Return(nil)
    d.tx.On("Rollback").Return(nil) // defer selalu dipanggil

    result, err := svc.CreateWallet(context.Background(), userID.String(), req)

    assert.NoError(t, err)
    assert.Equal(t, w.Name, result.Name)
    d.assertAll(t)
}
```

> **Aturan `Rollback`:** Selalu mock `Rollback` jika ada `defer tx.Rollback()` di kode. Meskipun `Commit` berhasil, `Rollback` tetap **terpanggil** di defer karena di GORM, rollback pada committed transaction tidak error.

---

## 7. Checklist Test Case per Method Type

### Query Method (Read-only)

| Skenario                 | Wajib                  |
| ------------------------ | ---------------------- |
| Success — data ditemukan | ✅                     |
| Success — list kosong    | ✅ (jika return slice) |
| Repository error         | ✅                     |
| Record not found         | ✅ (jika GetByID)      |

### Mutating Method dengan Transaction (Create / Update / Delete)

| Skenario                                              | Wajib         |
| ----------------------------------------------------- | ------------- |
| Success end-to-end                                    | ✅            |
| Input validation error (UUID tidak valid, dll.)       | ✅            |
| Lookup dependency gagal (misal: WalletType not found) | ✅            |
| Begin transaction error                               | ✅            |
| Insert/Update/Delete to DB error                      | ✅            |
| External call error (gRPC / HTTP)                     | ✅ (jika ada) |
| Outbox / side-effect create error                     | ✅ (jika ada) |
| Commit error                                          | ✅            |

### Worker / Background Job

| Skenario                                                          | Wajib |
| ----------------------------------------------------------------- | ----- |
| Constructor — verifikasi field default                            | ✅    |
| Context cancellation menghentikan loop                            | ✅    |
| Tidak ada data pending — tidak proses apa-apa                     | ✅    |
| Repository error — error di-wrap dan dikembalikan                 | ✅    |
| External client error — error di-log, tidak menghentikan proses   | ✅    |
| Retry counter increment gagal — di-log, lanjut ke message berikut | ✅    |
| Max retries terlampaui — di-log                                   | ✅    |

---

## 8. Aturan Penting yang Sering Terlewat

### 8.1 Scope Error di `defer`

```go
// DALAM SERVICE CODE:
func (s *service) CreateWallet(...) (dto.Response, error) {
    // ...
    tx, err := s.txManager.Begin(ctx)
    defer func() {
        tx.Rollback()
        if err != nil { // ← ini merujuk ke variabel `err` di scope function
            s.client.Cancel(ctx)
        }
    }()

    // ...

    // ← ini menggunakan `:=` (short variable), BUKAN `=`
    // sehingga `err` di scope defer TIDAK berubah
    if err := s.outboxRepo.Create(ctx, tx, msg); err != nil {
        return dto.Response{}, err
    }
}
```

**Implikasinya di test:** Jika outbox atau commit error menggunakan `if err :=` (bukan `err =`), maka `CancelInitialDeposit` di defer **TIDAK akan dipanggil**. Test harus merefleksikan ini:

```go
func TestCreateWallet_OutboxCreateError(t *testing.T) {
    // ...
    d.outboxRepo.On("Create", ...).Return(errors.New("outbox error"))
    d.tx.On("Rollback").Return(nil)

    // JANGAN mock CancelInitialDeposit karena tidak akan dipanggil
    result, err := svc.CreateWallet(...)

    assert.Error(t, err)
    d.txClient.AssertNotCalled(t, "CancelInitialDeposit") // ← eksplisit verifikasi tidak dipanggil
    d.assertAll(t)
}
```

### 8.2 Mock yang Return Pointer

```go
// SALAH — akan panic jika mock return nil
func (m *MockClient) Call(ctx context.Context) (*pb.Response, error) {
    args := m.Called(ctx)
    return args.Get(0).(*pb.Response), args.Error(1) // ← panic jika nil
}

// BENAR — cek nil dulu
func (m *MockClient) Call(ctx context.Context) (*pb.Response, error) {
    args := m.Called(ctx)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*pb.Response), args.Error(1)
}
```

### 8.3 `mock.Anything` vs Nilai Spesifik

```go
// Gunakan nilai spesifik jika ingin memverifikasi argumen yang dikirim
d.walletsRepo.On("GetWalletByID", mock.Anything, nil, id).Return(w, nil)
//                                ^ctx             ^tx  ^id spesifik

// Gunakan mock.Anything jika argumen berisi UUID yang di-generate di dalam service
d.walletsRepo.On("CreateWallet", mock.Anything, d.tx, mock.Anything).Return(w, nil)
//                                                                   ^tidak bisa prediksi UUID baru

// Gunakan mock.MatchedBy untuk validasi partial
d.walletsRepo.On("CreateWallet", mock.Anything, d.tx, mock.MatchedBy(func(w model.Wallets) bool {
    return w.Name == "BCA" && w.Balance > 0
})).Return(w, nil)
```

### 8.4 Background Goroutine — Gunakan `.Maybe()`

```go
func TestStart_ContextCancellation(t *testing.T) {
    publisher.interval = 10 * time.Millisecond // percepat interval

    // `.Maybe()` — mock ini boleh tidak dipanggil (tergantung timing)
    repo.On("GetPendingMessages", mock.Anything, publisher.batchSize).
        Return([]model.OutboxMessage{}, nil).Maybe()

    ctx, cancel := context.WithCancel(context.Background())
    done := make(chan struct{})

    go func() {
        publisher.Start(ctx)
        close(done)
    }()

    time.Sleep(50 * time.Millisecond)
    cancel()

    select {
    case <-done:
        // OK
    case <-time.After(2 * time.Second):
        t.Fatal("goroutine did not stop")
    }
}
```

### 8.5 Method yang Tidak Bisa Di-mock

Jika sebuah dependensi menggunakan concrete type (bukan interface), misalnya `*amqp091.Channel`, method-nya tidak bisa di-mock. Solusinya:

```go
func TestPublishMessage_CannotFullyMock(t *testing.T) {
    // Dokumentasikan kenapa test ini di-skip
    t.Skip("publishMessage requires a real amqp091.Channel; covered by integration tests")
}
```

Alternatif jangka panjang: bungkus concrete type di balik interface:

```go
type AMQPChannel interface {
    ExchangeDeclare(name, kind string, ...) error
    PublishWithContext(ctx context.Context, exchange, key string, ...) error
    Close() error
}
```

### 8.6 `AssertNotCalled` untuk Verifikasi Negative

```go
// Verifikasi bahwa method TIDAK dipanggil
d.txClient.AssertNotCalled(t, "InitialDeposit")
d.txClient.AssertNotCalled(t, "CancelInitialDeposit")
```

Gunakan ini untuk test seperti:

- `CreateWalletGRPC` dengan `balance == 0` tidak boleh memanggil `InitialDeposit`
- Error yang terjadi setelah `InitialDeposit` berhasil tidak selalu memanggil `CancelInitialDeposit`

---

## 9. Menjalankan Test & Cek Coverage

```bash
# Jalankan semua test di service layer
go test ./internal/service/... -v

# Dengan race detector (deteksi concurrent bug)
go test ./internal/service/... -v -race

# Coverage report
go test ./internal/service/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Coverage summary per function
go tool cover -func=coverage.out

# Minimum coverage gate (fail jika < 90%)
go test ./internal/service/... -covermode=atomic | grep -E "coverage: [0-9]+\.[0-9]+"
```

---

## 10. Referensi Cepat

### Mock Method Quick Reference

| Tujuan                     | Syntax                                                       |
| -------------------------- | ------------------------------------------------------------ |
| Mock return sukses         | `mock.On("Method", args...).Return(result, nil)`             |
| Mock return error          | `mock.On("Method", args...).Return(zero, errors.New("msg"))` |
| Mock boleh tidak dipanggil | `.Maybe()` di akhir                                          |
| Verifikasi semua mock      | `mock.AssertExpectations(t)`                                 |
| Verifikasi tidak dipanggil | `mock.AssertNotCalled(t, "MethodName")`                      |
| Match arg apapun           | `mock.Anything`                                              |
| Match tipe tertentu        | `mock.AnythingOfType("string")`                              |
| Match kondisi custom       | `mock.MatchedBy(func(v T) bool { ... })`                     |

### Assert Quick Reference

| Tujuan                     | Syntax                                       |
| -------------------------- | -------------------------------------------- |
| Tidak ada error            | `assert.NoError(t, err)`                     |
| Ada error                  | `assert.Error(t, err)`                       |
| Error mengandung substring | `assert.Contains(t, err.Error(), "keyword")` |
| Nilai sama                 | `assert.Equal(t, expected, actual)`          |
| Slice panjang N            | `assert.Len(t, slice, N)`                    |
| Slice kosong               | `assert.Empty(t, slice)`                     |
| Pointer nil                | `assert.Nil(t, val)`                         |
| Field kosong (zero value)  | `assert.Empty(t, result.ID)`                 |
| Field tidak kosong         | `assert.NotEmpty(t, result.ID)`              |

### Naming Convention

```
TestMethodName_Scenario

Contoh:
TestGetAllWallets_Success
TestGetAllWallets_EmptyList
TestGetAllWallets_RepositoryError
TestCreateWallet_InvalidUserID
TestCreateWallet_BeginTxError
TestCreateWallet_GRPCDepositError
TestStart_ContextCancellation
```
