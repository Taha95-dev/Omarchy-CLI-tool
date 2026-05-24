package database

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	MySQL      DatabaseType = "mysql"
	SQLite     DatabaseType = "sqlite"
)

type DBConfig struct {
	Type     DatabaseType
	Name     string
	User     string
	Password string
	Host     string
	Port     string
}

func DetectDatabase() DatabaseType {
	// Check go.mod for GORM or database drivers
	if _, err := os.Stat("go.mod"); err == nil {
		data, _ := os.ReadFile("go.mod")
		content := string(data)
		if strings.Contains(content, "gorm.io/driver/postgres") {
			return PostgreSQL
		}
		if strings.Contains(content, "gorm.io/driver/mysql") {
			return MySQL
		}
		if strings.Contains(content, "gorm.io/driver/sqlite") {
			return SQLite
		}
	}

	// Check package.json for Prisma, TypeORM
	if _, err := os.Stat("package.json"); err == nil {
		data, _ := os.ReadFile("package.json")
		content := string(data)
		if strings.Contains(content, "prisma") {
			return PostgreSQL // Prisma default
		}
		if strings.Contains(content, "typeorm") {
			return PostgreSQL
		}
	}

	return SQLite // Default for dev
}

func InitDatabase(dbType DatabaseType, projectName string) error {
	fmt.Printf("🐘 Initializing %s database for %s...\n", dbType, projectName)

	switch dbType {
	case PostgreSQL:
		return initPostgres(projectName)
	case MySQL:
		return initMySQL(projectName)
	case SQLite:
		return initSQLite(projectName)
	}
	return fmt.Errorf("unsupported database type: %s", dbType)
}

func initPostgres(projectName string) error {
	// Create docker-compose.yml for Postgres
	composeContent := fmt.Sprintf(`version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: %s_dev
      POSTGRES_USER: %s_user
      POSTGRES_PASSWORD: dev_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
`, projectName, projectName)

	err := os.WriteFile("docker-compose.yml", []byte(composeContent), 0644)
	if err != nil {
		return err
	}

	// Create .env file
	envContent := `DATABASE_URL=postgresql://%s_user:dev_password@localhost:5432/%s_dev?sslmode=disable`
	envContent = fmt.Sprintf(envContent, projectName, projectName)
	os.WriteFile(".env", []byte(envContent), 0644)

	fmt.Println("✅ Created docker-compose.yml")
	fmt.Println("✅ Created .env with DATABASE_URL")
	fmt.Println("\n📌 To start database:")
	fmt.Println("   docker-compose up -d")

	return nil
}

func initSQLite(projectName string) error {
	// Create data directory
	os.MkdirAll("data", 0755)

	// Create .env
	envContent := `DATABASE_URL=sqlite://./data/%s.db`
	envContent = fmt.Sprintf(envContent, projectName)
	os.WriteFile(".env", []byte(envContent), 0644)

	// Create empty database file
	dbPath := fmt.Sprintf("data/%s.db", projectName)
	f, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	f.Close()

	fmt.Printf("✅ Created SQLite database: %s\n", dbPath)
	fmt.Println("✅ Created .env with DATABASE_URL")

	return nil
}

func initMySQL(projectName string) error {
	// Similar to Postgres but with MySQL config
	composeContent := fmt.Sprintf(`version: '3.8'
services:
  mysql:
    image: mysql:8
    environment:
      MYSQL_DATABASE: %s_dev
      MYSQL_USER: %s_user
      MYSQL_PASSWORD: dev_password
      MYSQL_ROOT_PASSWORD: root_password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
`, projectName, projectName)

	err := os.WriteFile("docker-compose.yml", []byte(composeContent), 0644)
	if err != nil {
		return err
	}

	envContent := `DATABASE_URL=mysql://%s_user:dev_password@localhost:3306/%s_dev`
	envContent = fmt.Sprintf(envContent, projectName, projectName)
	os.WriteFile(".env", []byte(envContent), 0644)

	fmt.Println("✅ Created docker-compose.yml")
	fmt.Println("✅ Created .env with DATABASE_URL")

	return nil
}

func RunMigration(dbType DatabaseType, dryRun bool) error {
	if dryRun {
		fmt.Printf("🔍 DRY RUN: Would run migrations for %s\n", dbType)
		return previewMigration(dbType)
	}

	fmt.Printf("📦 Running migrations for %s...\n", dbType)

	fmt.Printf("📦 Running migrations for %s...\n", dbType)

	// Detect migration tool
	if _, err := os.Stat("prisma"); err == nil {
		fmt.Println("   Using Prisma...")
		cmd := exec.Command("npx", "prisma", "migrate", "dev", "--name", "init")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if _, err := os.Stat("migrations"); err == nil {
		fmt.Println("   Using golang-migrate...")
		// Assume migrate CLI is installed
		cmd := exec.Command("migrate", "-path", "migrations", "-database", "$DATABASE_URL", "up")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	fmt.Println("⚠️ No migration tool detected. Supported: Prisma, golang-migrate")
	fmt.Println("   Run 'omarchy db init' first to set up database")
	return nil
}
func previewMigration(dbType DatabaseType) error {
	fmt.Println("📋 Migration preview:")

	switch dbType {
	case PostgreSQL:
		// Show what would run
		fmt.Println("  CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(255) PRIMARY KEY);")
		fmt.Println("  -- Would run pending migration files")

	case SQLite:
		fmt.Println("  -- Would create/update tables based on schema")

	default:
		fmt.Println("  -- Preview not available for this database type")
	}

	// Check for migration files
	if _, err := os.Stat("migrations"); err == nil {
		files, _ := filepath.Glob("migrations/*.sql")
		fmt.Printf("  Found %d migration files\n", len(files))
		for _, f := range files {
			fmt.Printf("    - %s\n", filepath.Base(f))
		}
	} else if _, err := os.Stat("prisma/schema.prisma"); err == nil {
		fmt.Println("  Found Prisma schema")
		fmt.Println("  Would run: npx prisma migrate dev")
	} else {
		fmt.Println("  No migration files found")
		fmt.Println("  Run 'omarchy db init' first")
	}

	return nil
}
func SeedDatabase(dbType DatabaseType) error {
	fmt.Printf("🌱 Seeding %s database...\n", dbType)

	// Check for seed file
	if _, err := os.Stat("prisma/seed.ts"); err == nil {
		cmd := exec.Command("npx", "prisma", "db", "seed")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if _, err := os.Stat("seed.sql"); err == nil {
		// Run SQL seed file
		var cmd *exec.Cmd
		switch dbType {
		case PostgreSQL:
			cmd = exec.Command("psql", "-d", "$DATABASE_URL", "-f", "seed.sql")
		case MySQL:
			cmd = exec.Command("mysql", "--database", "$DATABASE_URL", "<", "seed.sql")
		default:
			fmt.Println("ℹ️ No seed file found. Create seed.sql or prisma/seed.ts")
			return nil
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	fmt.Println("ℹ️ No seed file found. Create seed.sql or prisma/seed.ts")
	return nil
}

func ResetDatabase(dbType DatabaseType) error {
	fmt.Printf("⚠️ This will DELETE all data in %s database!\n", dbType)
	fmt.Print("   Are you sure? (y/N): ")

	var resp string
	fmt.Scanln(&resp)
	if resp != "y" && resp != "Y" {
		fmt.Println("Aborted.")
		return nil
	}

	switch dbType {
	case PostgreSQL:
		cmd := exec.Command("docker-compose", "down", "-v")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		cmd = exec.Command("docker-compose", "up", "-d")
		return cmd.Run()

	case SQLite:
		// Delete .db file
		files, _ := filepath.Glob("data/*.db")
		for _, f := range files {
			os.Remove(f)
		}
		fmt.Println("✅ Database reset. Run 'omarchy db init' to recreate.")

	default:
		fmt.Println("Please manually reset your database")
	}

	return nil
}
