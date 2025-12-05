# Mire - High Performance Go Logging Library

Mire adalah high-performance logging library untuk Go yang dirancang untuk zero-allocation dan throughput tinggi. Proyek ini berfokus pada performa, keamanan, dan fleksibilitas.

## Fitur Utama

1. **Zero-Allocation Design**: Menggunakan `[]byte` alih-alih `string` dan manipulasi byte manual untuk menghindari alokasi memori
2. **Extensive Object Pooling**: Menggunakan `sync.Pool` dan goroutine-local pools untuk mengurangi pressure GC
3. **Asynchronous Logging**: Mendukung logging non-blocking untuk throughput tinggi
4. **Berbagai Formatter**: Text, JSON, dan CSV formatter dengan zero-allocation
5. **Context-Aware Logging**: Otomatis mengekstrak informasi dari context (trace_id, user_id, dll.)
6. **Hook System**: Ekstensibilitas untuk pemrosesan log tambahan
7. **Sensitive Data Masking**: Perlindungan otomatis untuk data sensitif
8. **Distributed Tracing Support**: Integrasi dengan sistem tracing terdistribusi
9. **Log Sampling**: Rate limiting untuk volume log tinggi
10. **Thread-Safe**: Aman untuk digunakan secara konkuren antar goroutine

## Struktur Proyek

```
mire/
├── core/           # Definisi struktur data dasar
├── logger/         # Implementasi fungsionalitas logging utama  
├── formatter/      # Modul untuk memformat output log
├── util/           # Fungsionalitas bantu dan pooling
├── hook/           # Sistem untuk ekstensibilitas
├── writer/         # Modul untuk penulisan log
├── metric/         # Kolektor metrik
├── config/         # Definisi konfigurasi
├── sampler/        # Log sampling
├── errors/         # Definisi error khusus
├── example/        # Contoh penggunaan
├── main.go         # Contoh aplikasi
├── go.mod          # Definisi module
└── README.md       # Dokumentasi utama
```

## Arsitektur Kunci

### 1. Zero-Allocation Pooling System
- Gunakan `[]byte` alih-alih `string`
- `sync.Pool` untuk buffer, map, dan struktur data lainnya
- Goroutine-local pools untuk mengurangi lock contention
- Object reuse secara menyeluruh

### 2. Sistem Formatter yang Efisien
- Manual byte manipulation tanpa fmt
- Zero-allocation formatting
- Optimisasi untuk berbagai kebutuhan (Text, JSON, CSV)

### 3. Clock Internal
- Atomic time updates untuk efisiensi timestamp
- Clock internal untuk menghindari overhead pembuatan time.Now()

### 4. Asynchronous Processing
- Non-blocking logging untuk throughput tinggi
- Multiple worker goroutines
- Buffer channels untuk manajemen backpressure

## Cara Penggunaan

### Logger Dasar
```go
import "github.com/Lunar-Chipter/mire/logger"

// Logger default dengan konfigurasi standar
log := logger.NewDefaultLogger()
defer log.Close()

log.Info("Aplikasi dimulai")
log.WithFields(map[string]interface{}{
    "user_id": 123,
    "action": "login",
}).Info("User berhasil login")
```

### Logger dengan Konfigurasi Kustom
```go
config := logger.LoggerConfig{
    Level:   core.INFO,
    Output:  os.Stdout,
    Formatter: &formatter.JSONFormatter{
        PrettyPrint:       false,
        ShowTimestamp:     true,
        EnableStackTrace:  true,
    },
    AsyncLogging:      true,
    AsyncWorkerCount:  4,
}

log := logger.New(config)
defer log.Close()
```

### Context-Aware Logging
```go
ctx := context.Background()
ctx = util.WithTraceID(ctx, "trace-123")
ctx = util.WithUserID(ctx, "user-456")

log.InfoC(ctx, "Processing request") // Akan menyertakan trace_id dan user_id
```

## Unit Test dan Benchmark

Proyek ini dilengkapi dengan berbagai unit test dan benchmark:

- **Core Tests**: Pengujian operasi pool, serialisasi, metrik
- **Logger Tests**: Pengujian fungsionalitas logging, level filtering, concurrency
- **Formatter Tests**: Pengujian berbagai formatter dan fitur-fitur mereka
- **Benchmark Tests**: Evaluasi kinerja berbagai komponen

## Best Practices

1. **Tutup logger saat selesai**:
   ```go
   defer log.Close()
   ```

2. **Gunakan async logging untuk volume tinggi**:
   ```go
   AsyncLogging: true
   AsyncWorkerCount: 4
   ```

3. **Gunakan formatter yang sesuai kebutuhan**:
   - CSV untuk performa tertinggi
   - JSON untuk log struktural
   - Text untuk debugging
  
4. **Gunakan level log secara tepat** sesuai kebutuhan aplikasi

5. **Manfaatkan context untuk distributed tracing**:
   ```go
   log.InfoC(ctx, "message")  // Otomatis menyertakan trace info dari context
   ```

## Performa

Mire dirancang untuk performa tinggi:
- CSV Formatter Batch: 60 juta+ operasi/detik (~24ns/op) dengan 0 alokasi
- Throughput tinggi dengan minimal alokasi memori
- Efisiensi CPU tinggi berkat pengoptimalan caching dan pooling

## Lisensi

Lisensi Apache 2.0 - silakan lihat file LICENSE untuk detail lengkapnya.

---

Proyek ini siap untuk produksi dan digunakan di lingkungan dengan kebutuhan performa tinggi.