package main

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/logger"
	"github.com/Lunar-Chipter/mire/util"
	// "github.com/Lunar-Chipter/mire/hook" // Removed import for hook package
)

// wrappedError wraps an error with a message
type wrappedError struct {
	msg   string
	cause error
}

func (e *wrappedError) Error() string {
	if e.cause != nil {
		return e.msg + ": " + e.cause.Error()
	}
	return e.msg
}

func (e *wrappedError) Unwrap() error {
	return e.cause
}

// printLine is a helper function to print lines without fmt
func printLine(s string) {
	os.Stdout.Write([]byte(s))
	os.Stdout.Write([]byte("\n"))
}

func main() {
	printLine("===================================================")
	printLine("  DEMONSTRASI PENGGUNAAN LIBRARY LOGGING MIRE      ")
	printLine("===================================================")
	printLine("File ini menunjukkan berbagai cara menggunakan logger.")
	printLine("Perhatikan output di konsol, app.log, dan errors.log.")
	printLine("---------------------------------------------------")

	// --- Contoh 1: Logger Default (Console Output) ---
	// Logger default sudah dikonfigurasi untuk menampilkan warna, timestamp,
	// info pemanggil, dan level INFO ke atas.
	printLine("### 1. Logger Default (Output Konsol Lebih Ringkas) ###")
	// Konfigurasi logger default untuk mengurangi output ke konsol
	defaultConfig := logger.LoggerConfig{
		Level:             core.WARN, // Hanya tampilkan WARN ke atas di konsol
		Output:            os.Stdout,
		ErrorOutput:       io.Discard, // Buang pesan error internal logger
		CallerDepth:       logger.DEFAULT_CALLER_DEPTH,
		TimestampFormat:   logger.DEFAULT_TIMESTAMP_FORMAT,
		BufferSize:        logger.DEFAULT_BUFFER_SIZE,
		FlushInterval:     logger.DEFAULT_FLUSH_INTERVAL,
		AsyncWorkerCount:  4,
		ClockInterval: 10 * time.Millisecond,
		Formatter: &formatter.TextFormatter{
			EnableColors:      true,
			ShowTimestamp:     true,
			ShowCaller:        true,
			TimestampFormat:   logger.DEFAULT_TIMESTAMP_FORMAT,
		},
	}
	logDefault := logger.New(defaultConfig)
	defer logDefault.Close() // Pastikan logger ditutup dengan bersih.

	logDefault.Info("Ini adalah pesan INFORMASI dari logger default. (Tidak akan muncul di konsol karena Level WARN).")
	logDefault.Warnf("Ada %d peringatan di sistem.", 2)
	logDefault.Debug("Pesan debug ini TIDAK akan muncul karena level default adalah WARN.") // Tidak akan muncul
	logDefault.Trace("Pesan trace ini juga TIDAK akan muncul.")                              // Tidak akan muncul
	logDefault.Error("Terjadi error sederhana.")
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond) // Memberi waktu untuk flush buffer jika ada

	// --- Contoh 2: Logger dengan Fields dan Context ---
	// Logger memungkinkan penambahan bidang (key-value pairs) ke setiap entri log.
	// Kita juga bisa menambahkan informasi kontekstual seperti TraceID, SpanID.
	printLine("### 2. Logger dengan Fields & Context ###")
	ctx := context.Background()
	// Menambahkan TraceID, SpanID, UserID ke konteks.
	// Ini akan diekstrak otomatis oleh logger jika diaktifkan.
	ctx = util.WithTraceID(ctx, "trace-xyz-987")
	ctx = util.WithSpanID(ctx, "span-123")
	ctx = util.WithUserID(ctx, "user-alice")

	// Menggunakan logger default untuk menunjukkan field dan konteks.
	logWithContext := logDefault.WithFieldsBytes(map[string][]byte{
		"service": core.StringToBytes("auth-service"),
		"version": core.StringToBytes("1.0.0"),
	})
	logWithContext.Info("Pengguna berhasil login.",
		"username", "alice",
		"ip_address", "192.168.1.100")

	// Log dengan konteks eksplisit menggunakan metode context-aware.
	logWithContext.InfofC(ctx, "Memproses permintaan otorisasi untuk %s.", "token-ABC")
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// --- Contoh 3: Error Logging dengan Stack Trace ---
	// Logger dapat merekam error dan menyertakan stack trace untuk debugging.
	printLine("### 3. Error Logging dengan Stack Trace ###")
	errSample := errors.New("gagal membaca konfigurasi database")
	logDefault.WithFields(map[string]interface{}{
		"error_code": 500,
		"component":  "database-connector",
	}).Error("Error saat inisialisasi:", errSample.Error())
	// Default logger sudah mengaktifkan ShowStackTrace untuk level ERROR ke atas.
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// --- Contoh 4: Logger JSON ke File (app.log) ---
	// Mengonfigurasi logger untuk menulis log dalam format JSON ke file.
	printLine("### 4. Logger JSON ke File (app.log) ###")
	printLine("Log JSON akan ditulis ke 'app.log'. Periksa isinya setelah program selesai.")
	jsonFileLogger, err := setupJSONFileLogger("app.log")
	if err != nil {
		logDefault.Fatalf("Failed to setup JSON file logger: %v", err) // Use logDefault to fatal here
	}
	defer jsonFileLogger.Close() // Penting: Tutup logger agar buffer di-flush ke file!

	jsonFileLogger.Debug("Pesan debug untuk JSON file logger.")
	jsonFileLogger.WithFields(map[string]interface{}{
		"trans_id": "TXN-001",
		"amount":   123.45,
		"currency": "IDR",
	}).Info("Transaksi berhasil diproses.")
	jsonFileLogger.Error("Gagal menyimpan data pengguna ke cache.",
		"user_id", "user-bob",
		"cache_key", "user:bob")
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// --- Contoh 5: Custom Text Logger (Tanpa Timestamp & Caller) ---
	// Membuat logger dengan format teks kustom, menyembunyikan beberapa metadata.
	printLine("### 5. Custom Text Logger ###")
	customTextLogger := setupCustomTextLogger()
	customTextLogger.Notice("Ini adalah pesan 'NOTICE' dari logger kustom (tanpa timestamp/caller).")
	customTextLogger.Infof("Level terendah: %s", core.TRACE.String())
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// --- Contoh 6: Demonstrasi Hooks (errors.log) ---
	// Menunjukkan cara mengonfigurasi dan menggunakan hook.
	// Log level ERROR ke atas akan ditulis ke 'errors.log'.
	// printLine("### 6. Demonstrasi Hooks ###") // Commented out
	// demonstrateHooks() // Commented out
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// Pastikan semua buffer di-flush sebelum program berakhir.
	// Terutama penting untuk buffered writer dan async logger.
	// logger.NewDefaultLogger().Close() // Jika NewDefaultLogger dipanggil berkali-kali, hanya perlu menutup instance yang digunakan.
	// jsonFileLogger.Close() // Pastikan ditutup jika belum di-defer
	// Biasanya, main logger akan ditutup pada akhir aplikasi.
	// Untuk demo ini, kita tidak menutup logDefault secara eksplisit di sini,
	// karena ia langsung menulis ke os.Stdout, tetapi jika ia memiliki buffered writer,
	// maka harus ditutup.

	printLine("===================================================")
	printLine("  DEMONSTRASI SELESAI                              ")
	printLine("===================================================")
	printLine("Periksa file 'app.log' dan 'errors.log' untuk melihat output log.")
}

func setupJSONFileLogger(filePath string) (*logger.Logger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		return nil, &wrappedError{
			msg:   "gagal membuka file log " + filePath,
			cause: err,
		}
	}

	jsonConfig := logger.LoggerConfig{
		Level:       core.DEBUG,
		Output:      file,
		ErrorOutput: io.Discard, // Buang pesan error internal logger
		BufferSize:  1024, // Aktifkan buffered writer
		Formatter: &formatter.JSONFormatter{
			PrettyPrint:      true,
			ShowCaller:       true,
			EnableStackTrace: true,
			TimestampFormat:  logger.DEFAULT_TIMESTAMP_FORMAT,
		},
	}
	return logger.New(jsonConfig), nil
}

// setupCustomTextLogger membuat logger dengan format teks yang disederhanakan.
func setupCustomTextLogger() *logger.Logger {
	customConfig := logger.LoggerConfig{
		Level:             core.TRACE, // Tampilkan semua log, bahkan trace
		ErrorOutput:       io.Discard, // Buang pesan error internal logger
		Formatter: &formatter.TextFormatter{
			EnableColors:      true,
			ShowTimestamp:     false, // Sembunyikan timestamp
			ShowCaller:        false, // Sembunyikan info pemanggil
			ShowPID:           true,  // Tampilkan Process ID
			ShowGoroutine:     true,  // Tampilkan Goroutine ID
		},
	}
	return logger.New(customConfig)
}

// demonstrateHooks menunjukkan cara mengonfigurasi dan menggunakan hook.
// Hook ini akan menulis semua log level ERROR atau di atasnya ke file 'errors.log'.
// func demonstrateHooks() {
// 	fmt.Println("Log level ERROR dan di atasnya akan ditulis ke 'errors.log'.")

// 	// Create a file hook for errors.log
// 	errorFileHook, err := hook.NewFileHook("errors.log") // Use mire/hook/NewFileHook
// 	if err != nil {
// 		fmt.Printf("Failed to create error file hook: %v\n", err)
// 		return
// 	}
// 	defer errorFileHook.Close() // Ensure the file hook is closed

// 	// 2. Konfigurasikan logger untuk menggunakan hook.
// 	logWithHook := logger.New(logger.LoggerConfig{
// 		Level:       core.INFO, // Logger ini akan menampilkan INFO ke atas ke konsol
// 		Output:      os.Stdout, // Console output
// 		ErrorOutput: io.Discard, // Discard internal logger errors
// 		EnableErrorFileHook: false, // Disable the built-in error file hook
// 		Hooks: []hook.Hook{
// 			errorFileHook, // Add the manual file hook
// 		},
// 		Formatter: &formatter.TextFormatter{ // This formatter is for console output
// 			EnableColors:    true,
// 			TimestampFormat: logger.DEFAULT_TIMESTAMP_FORMAT,
// 			ShowCaller:      true,
// 		},
// 	})
// 	defer logWithHook.Close() // Ensure this logger instance is closed

// 	// 3. Gunakan logger seperti biasa.
// 	logWithHook.Info("Ini pesan INFO, akan tampil di konsol, tapi tidak di 'errors.log'.")
// 	logWithHook.Warn("Ini pesan WARN, akan tampil di konsol, tapi tidak di 'errors.log'.")

// 	// Pesan ini akan masuk ke konsol DAN ke file errors.log karena hook.
// 	logWithHook.WithFields(map[string]interface{}{"db_host": "10.0.0.5"}).Error("Gagal menghubungkan ke database.")

// 	// Pesan ini juga akan memicu hook.
// 	logWithHook.WithFields(map[string]interface{}{"service": "payment-gateway"}).Error("Timeout transaksi.")
// }