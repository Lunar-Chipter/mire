# Mire - Usage Guide

Mire adalah high-performance logging library untuk Go yang dirancang untuk zero-allocation dan throughput tinggi. Dokumen ini menjelaskan cara menggunakan library ini secara efektif.

## Instalasi

```bash
go get github.com/Lunar-Chipter/mire
```

## Penggunaan Dasar

### 1. Logger Default

```go
import (
    "github.com/Lunar-Chipter/mire/logger"
)

// Membuat logger dengan konfigurasi default
log := logger.NewDefaultLogger()
defer log.Close()

// Logging dasar
log.Info("Aplikasi dimulai")
log.Warn("Ini adalah peringatan")
log.Error("Terjadi kesalahan")
```

### 2. Logger dengan Konfigurasi Kustom

```go
config := logger.LoggerConfig{
    Level:   core.INFO,
    Output:  os.Stdout,
    Formatter: &formatter.JSONFormatter{
        PrettyPrint:     true,
        ShowTimestamp:   true,
        ShowCaller:      true,
        EnableStackTrace: true,
    },
}

jsonLogger := logger.New(config)
defer jsonLogger.Close()
```

### 3. Logging dengan Fields

```go
log.WithFields(map[string]interface{}{
    "user_id": 123,
    "action":  "login",
}).Info("User berhasil login")
```

### 4. Logging Asinkron untuk Kinerja Tinggi

```go
asyncLogger := logger.New(logger.LoggerConfig{
    Level:                core.INFO,
    Output:              os.Stdout,
    AsyncLogging:        true,
    AsyncWorkerCount:    4,
    AsyncLogChannelBufferSize: 2000,
})
defer asyncLogger.Close()

// Operasi logging tidak akan memblokir thread utama
for i := 0; i < 1000; i++ {
    asyncLogger.WithFields(map[string]interface{}{
        "iteration": i,
    }).Info("Pesan log asinkron")
}
```

## Fitur Lanjutan

### 1. Formatter

Mire mendukung tiga jenis formatter:

#### JSON Formatter
```go
jsonFormatter := &formatter.JSONFormatter{
    ShowCaller:        true,
    ShowTraceInfo:     true,
    EnableStackTrace:  true,
    SensitiveFields:   []string{"password", "token"},
    MaskSensitiveData: true,
}
```

#### Text Formatter
```go
textFormatter := &formatter.TextFormatter{
    EnableColors:      true,
    ShowTimestamp:     true,
    ShowCaller:        true,
    EnableStackTrace:  true,
    SensitiveFields:   []string{"password", "token"},
    MaskSensitiveData: true,
}
```

#### CSV Formatter
```go
csvFormatter := &formatter.CSVFormatter{
    IncludeHeader:     true,
    FieldOrder:        []string{"timestamp", "level", "message", "user_id", "action"},
    SensitiveFields:   []string{"password", "token"},
    MaskSensitiveData: true,
}
```

### 2. Context-Aware Logging

```go
// Menambahkan informasi konteks
ctx := context.Background()
ctx = util.WithTraceID(ctx, "trace-123")
ctx = util.WithUserID(ctx, "user-456")

// Logger akan otomatis mengekstrak informasi dari konteks
log.InfoC(ctx, "Memproses permintaan")
```

### 3. Sistem Hook

```go
// Implementasi hook kustom
type CustomHook struct {
    endpoint string
}

func (h *CustomHook) Fire(entry *core.LogEntry) error {
    // Kirim log entry ke layanan eksternal
    payload, err := json.Marshal(entry)
    if err != nil {
        return err
    }

    resp, err := http.Post(h.endpoint, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (h *CustomHook) Close() error {
    // Membersihkan resource
    return nil
}

// Menggunakan hook kustom
customHook := &CustomHook{endpoint: "https://logs.example.com/api"}
log := logger.New(logger.LoggerConfig{
    Level: core.INFO,
    Output: os.Stdout,
    Hooks: []hook.Hook{customHook},
})
```

## Konfigurasi Kinerja

### 1. Optimasi untuk Throughput Tinggi

```go
highPerfConfig := logger.LoggerConfig{
    AsyncLogging:        true,              // Gunakan logging asinkron
    AsyncWorkerCount:    8,                 // Jumlah worker untuk async logging
    AsyncLogChannelBufferSize: 5000,        // Buffer untuk channel async
    BufferSize:          4096,              // Ukuran buffer untuk buffered writer
    FlushInterval:       100 * time.Millisecond, // Interval flush
    DisableLocking:      true,              // Nonaktifkan internal locking jika aman
}
```

### 2. Optimasi untuk Zero-Allocation

Mire dirancang untuk zero-allocation logging sebanyak mungkin. Untuk memaksimalkan efisiensi:

- Gunakan `[]byte` alih-alih `string` saat mengirim data ke logger
- Gunakan pooling object melalui fungsi yang disediakan
- Gunakan formatter yang sesuai kebutuhan (CSV formatter memiliki performa tertinggi)

## Keamanan dan Penanganan Error

### 1. Masking Data Sensitif

```go
jsonFormatter := &formatter.JSONFormatter{
    SensitiveFields:   []string{"password", "token", "credit_card"},
    MaskSensitiveData: true,
    MaskStringValue:   "[MASKED]", // Nilai yang digunakan untuk masking
}
```

### 2. Penanganan Error Aman

```go
// Mire dirancang dengan prinsip "Never let logging crash your application"
// Logger akan tetap berjalan bahkan jika terjadi error internal
config := logger.LoggerConfig{
    ErrorHandler: func(err error) {
        // Tangani error internal logger
        fmt.Printf("Logger error: %v\n", err)
    },
}
```

## Best Practices

1. **Tutup logger saat selesai**:
   Gunakan `defer log.Close()` untuk memastikan semua log yang tertunda ditulis sebelum aplikasi berhenti.

2. **Gunakan async logging untuk volume tinggi**:
   Jika aplikasi Anda menghasilkan banyak log, aktifkan async logging untuk menghindari blocking pada jalur utama aplikasi.

3. **Gunakan pooling secara efektif**:
   Mire memiliki sistem pooling objek yang luas; implementasi yang benar akan mengurangi tekanan GC secara signifikan.

4. **Gunakan level log secara tepat**:
   - `TRACE`: Informasi debugging sangat detail
   - `DEBUG`: Informasi debugging
   - `INFO`: Pesan informasi umum
   - `NOTICE`: Kondisi normal tapi signifikan
   - `WARN`: Peringatan
   - `ERROR`: Kesalahan
   - `FATAL`: Kesalahan kritis yang menyebabkan terminasi
   - `PANIC`: Kondisi panic

5. **Optimalkan formatter berdasarkan kebutuhan**:
   - Gunakan CSV formatter untuk performa tertinggi
   - Gunakan JSON formatter untuk log yang akan diproses secara struktural
   - Gunakan Text formatter untuk debugging di lingkungan development

## Contoh Lengkap

```go
package main

import (
    "context"
    "os"
    "time"
    
    "github.com/Lunar-Chipter/mire/core"
    "github.com/Lunar-Chipter/mire/formatter"
    "github.com/Lunar-Chipter/mire/logger"
    "github.com/Lunar-Chipter/mire/util"
)

func main() {
    // Buat logger performa tinggi dengan async logging
    log := logger.New(logger.LoggerConfig{
        Level:   core.INFO,
        Output:  os.Stdout,
        Formatter: &formatter.TextFormatter{
            EnableColors:    true,
            ShowTimestamp:   true,
            ShowCaller:      true,
            ShowTraceInfo:   true,
        },
        AsyncLogging:        true,
        AsyncWorkerCount:    4,
        AsyncLogChannelBufferSize: 2000,
    })
    defer log.Close()

    // Tambahkan informasi konteks
    ctx := context.Background()
    ctx = util.WithTraceID(ctx, "trace-12345")
    ctx = util.WithUserID(ctx, "user-67890")

    // Logging dengan fields
    log.WithFields(map[string]interface{}{
        "user_id": 12345,
        "action":  "purchase",
        "amount":  99.99,
    }).Info("Transaksi selesai")

    // Logging dengan konteks
    log.InfoC(ctx, "Memproses permintaan") // Akan otomatis menyertakan trace_id dan user_id
}
```