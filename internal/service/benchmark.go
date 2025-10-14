package service

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

func RunBenchmark(db *sql.DB) {
	// Создаём папку для результатов
	os.MkdirAll("results", 0755)

	// === ЗАДАНИЕ 1: Партицирование ===
	runPartitionBenchmark(db)

	// === ЗАДАНИЕ 2: LOGGED vs UNLOGGED ===
	runUnloggedBenchmark(db)

	fmt.Println("✅ Все бенчмарки завершены. Результаты в папке results/")
}

// Задание 1: Партицирование
func runPartitionBenchmark(db *sql.DB) {
	file, _ := os.Create("results/partition_benchmark.txt")
	defer file.Close()

	fmt.Fprintln(file, "=== ЗАДАНИЕ 1: Партицирование ===")
	fmt.Fprintln(file, "Условие: attendance_logs — самая большая таблица")
	fmt.Fprintln(file, "Партицирование: по start_time (месячные партиции)")
	fmt.Fprintln(file, "")

	// По ключу партицирования
	start := time.Now()
	var count int
	db.QueryRow("SELECT COUNT(*) FROM attendance_logs_part WHERE start_time >= $1", "2024-05-01").Scan(&count)
	partTime := time.Since(start)

	start = time.Now()
	db.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE start_time >= $1", "2024-05-01").Scan(&count)
	origTime := time.Since(start)

	fmt.Fprintf(file, "Запрос: SELECT COUNT(*) WHERE start_time >= '2024-05-01'\n")
	fmt.Fprintf(file, "  Partitioned: %v\n", partTime)
	fmt.Fprintf(file, "  Original:    %v\n\n", origTime)

	// Не по ключу
	start = time.Now()
	db.QueryRow("SELECT COUNT(*) FROM attendance_logs_part WHERE user_id = $1", 12345).Scan(&count)
	partUserTime := time.Since(start)

	start = time.Now()
	db.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE user_id = $1", 12345).Scan(&count)
	origUserTime := time.Since(start)

	fmt.Fprintf(file, "Запрос: SELECT COUNT(*) WHERE user_id = 12345\n")
	fmt.Fprintf(file, "  Partitioned: %v\n", partUserTime)
	fmt.Fprintf(file, "  Original:    %v\n", origUserTime)
}

// Задание 2: LOGGED vs UNLOGGED
func runUnloggedBenchmark(db *sql.DB) {
	file, _ := os.Create("results/unlogged_benchmark.txt")
	defer file.Close()

	fmt.Fprintln(file, "=== ЗАДАНИЕ 2: LOGGED vs UNLOGGED ===")
	fmt.Fprintln(file, "Тесты: вставка (группой/по одной), чтение, удаление")
	fmt.Fprintln(file, "Объём данных: 10 000 записей")
	fmt.Fprintln(file, "")

	// Очистка
	db.Exec("DROP TABLE IF EXISTS test_logged, test_unlogged")

	// Создание
	db.Exec("CREATE TABLE test_logged (id SERIAL, data TEXT)")
	db.Exec("CREATE UNLOGGED TABLE test_unlogged (id SERIAL, data TEXT)")

	// Вставка группой (10k)
	start := time.Now()
	db.Exec("INSERT INTO test_unlogged SELECT generate_series(1,10000), 'data'")
	unloggedBulk := time.Since(start)

	start = time.Now()
	db.Exec("INSERT INTO test_logged SELECT generate_series(1,10000), 'data'")
	loggedBulk := time.Since(start)

	// Вставка по одной (1000 раз)
	start = time.Now()
	for i := 1; i <= 1000; i++ {
		db.Exec("INSERT INTO test_unlogged (data) VALUES ('single')")
	}
	unloggedSingle := time.Since(start)

	start = time.Now()
	for i := 1; i <= 1000; i++ {
		db.Exec("INSERT INTO test_logged (data) VALUES ('single')")
	}
	loggedSingle := time.Since(start)

	// Чтение
	start = time.Now()
	var count int
	db.QueryRow("SELECT COUNT(*) FROM test_unlogged").Scan(&count)
	unloggedRead := time.Since(start)

	start = time.Now()
	db.QueryRow("SELECT COUNT(*) FROM test_logged").Scan(&count)
	loggedRead := time.Since(start)

	// Удаление
	start = time.Now()
	db.Exec("DELETE FROM test_unlogged")
	unloggedDelete := time.Since(start)

	start = time.Now()
	db.Exec("DELETE FROM test_logged")
	loggedDelete := time.Since(start)

	// Запись в файл
	fmt.Fprintf(file, "Вставка 10k строк группой:\n")
	fmt.Fprintf(file, "  UNLOGGED: %v\n", unloggedBulk)
	fmt.Fprintf(file, "  LOGGED:   %v\n\n", loggedBulk)

	fmt.Fprintf(file, "Вставка 1k строк по одной:\n")
	fmt.Fprintf(file, "  UNLOGGED: %v\n", unloggedSingle)
	fmt.Fprintf(file, "  LOGGED:   %v\n\n", loggedSingle)

	fmt.Fprintf(file, "Чтение (COUNT):\n")
	fmt.Fprintf(file, "  UNLOGGED: %v\n", unloggedRead)
	fmt.Fprintf(file, "  LOGGED:   %v\n\n", loggedRead)

	fmt.Fprintf(file, "Удаление всех строк:\n")
	fmt.Fprintf(file, "  UNLOGGED: %v\n", unloggedDelete)
	fmt.Fprintf(file, "  LOGGED:   %v\n", loggedDelete)
}
