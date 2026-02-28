package deploy_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// projectRoot returns the absolute path to the project root (two levels up from this test file).
func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}
	// internal/deploy/deploy_test.go → project root is ../../
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

// --- Dockerfile.api ---

func TestDockerfileAPI_Exists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "Dockerfile.api")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Dockerfile.api does not exist")
	}
}

func TestDockerfileAPI_MultiStageBuild(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.api"))

	if !strings.Contains(content, "AS builder") {
		t.Error("Dockerfile.api should have a named builder stage")
	}
	if strings.Count(content, "FROM ") < 2 {
		t.Error("Dockerfile.api should be multi-stage (at least 2 FROM directives)")
	}
}

func TestDockerfileAPI_BuildsGoModule(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.api"))

	if !strings.Contains(content, "go mod download") {
		t.Error("Dockerfile.api should run go mod download")
	}
	if !strings.Contains(content, "go build") {
		t.Error("Dockerfile.api should run go build")
	}
	if !strings.Contains(content, "./cmd/api") {
		t.Error("Dockerfile.api should build ./cmd/api")
	}
}

func TestDockerfileAPI_CGODisabled(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.api"))

	if !strings.Contains(content, "CGO_ENABLED=0") {
		t.Error("Dockerfile.api should set CGO_ENABLED=0 for static binary")
	}
}

func TestDockerfileAPI_CopiesMigrations(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.api"))

	if !strings.Contains(content, "COPY migrations/") {
		t.Error("Dockerfile.api should copy migrations directory into runtime image")
	}
}

func TestDockerfileAPI_ExposesPort(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.api"))

	if !strings.Contains(content, "EXPOSE 8080") {
		t.Error("Dockerfile.api should expose port 8080")
	}
}

func TestDockerfileAPI_HasCACerts(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.api"))

	if !strings.Contains(content, "ca-certificates") {
		t.Error("Dockerfile.api runtime stage should install ca-certificates (needed for TLS to Resend/Telegram)")
	}
}

// --- Dockerfile.frontend ---

func TestDockerfileFrontend_Exists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "Dockerfile.frontend")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Dockerfile.frontend does not exist")
	}
}

func TestDockerfileFrontend_MultiStageBuild(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.frontend"))

	if !strings.Contains(content, "AS builder") {
		t.Error("Dockerfile.frontend should have a named builder stage")
	}
	if strings.Count(content, "FROM ") < 2 {
		t.Error("Dockerfile.frontend should be multi-stage (at least 2 FROM directives)")
	}
}

func TestDockerfileFrontend_UsesNpmCi(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.frontend"))

	if !strings.Contains(content, "npm ci") {
		t.Error("Dockerfile.frontend should use npm ci for reproducible installs")
	}
}

func TestDockerfileFrontend_BuildsFrontend(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.frontend"))

	if !strings.Contains(content, "npm run build") {
		t.Error("Dockerfile.frontend should run npm run build")
	}
}

func TestDockerfileFrontend_UsesNginxRuntime(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.frontend"))

	if !strings.Contains(content, "nginx") {
		t.Error("Dockerfile.frontend runtime stage should use nginx")
	}
}

func TestDockerfileFrontend_CopiesDist(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.frontend"))

	if !strings.Contains(content, "/usr/share/nginx/html") {
		t.Error("Dockerfile.frontend should copy built files to nginx html directory")
	}
}

func TestDockerfileFrontend_CopiesNginxConf(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "Dockerfile.frontend"))

	if !strings.Contains(content, "nginx.conf") {
		t.Error("Dockerfile.frontend should copy custom nginx.conf")
	}
}

// --- nginx.conf ---

func TestNginxConf_Exists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "frontend", "nginx.conf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("frontend/nginx.conf does not exist")
	}
}

func TestNginxConf_SPAFallback(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "frontend", "nginx.conf"))

	if !strings.Contains(content, "try_files") {
		t.Error("nginx.conf should use try_files for SPA fallback")
	}
	if !strings.Contains(content, "/index.html") {
		t.Error("nginx.conf should fall back to /index.html")
	}
}

func TestNginxConf_APIProxy(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "frontend", "nginx.conf"))

	if !strings.Contains(content, "location /api/") {
		t.Error("nginx.conf should have a location block for /api/")
	}
	if !strings.Contains(content, "proxy_pass http://api:8080") {
		t.Error("nginx.conf should proxy /api/ requests to the api service on port 8080")
	}
}

func TestNginxConf_ListensOnPort80(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "frontend", "nginx.conf"))

	if !strings.Contains(content, "listen 80") {
		t.Error("nginx.conf should listen on port 80")
	}
}

// --- docker-compose.yml ---

func TestDockerCompose_Exists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "docker-compose.yml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("docker-compose.yml does not exist")
	}
}

func TestDockerCompose_HasThreeServices(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	for _, svc := range []string{"api:", "frontend:", "postgres:"} {
		if !strings.Contains(content, svc) {
			t.Errorf("docker-compose.yml should define %s service", strings.TrimSuffix(svc, ":"))
		}
	}
}

func TestDockerCompose_APIUsesDockerfileAPI(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	if !strings.Contains(content, "Dockerfile.api") {
		t.Error("docker-compose.yml api service should reference Dockerfile.api")
	}
}

func TestDockerCompose_FrontendUsesDockerfileFrontend(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	if !strings.Contains(content, "Dockerfile.frontend") {
		t.Error("docker-compose.yml frontend service should reference Dockerfile.frontend")
	}
}

func TestDockerCompose_PostgresImage(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	if !strings.Contains(content, "postgres:16-alpine") {
		t.Error("docker-compose.yml should use postgres:16-alpine image")
	}
}

func TestDockerCompose_PostgresHealthcheck(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	if !strings.Contains(content, "pg_isready") {
		t.Error("docker-compose.yml postgres should have healthcheck using pg_isready")
	}
}

func TestDockerCompose_APIDependsOnPostgresHealthy(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	if !strings.Contains(content, "service_healthy") {
		t.Error("docker-compose.yml api service should depend on postgres with service_healthy condition")
	}
}

func TestDockerCompose_PersistentVolume(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	if !strings.Contains(content, "pgdata") {
		t.Error("docker-compose.yml should define a pgdata volume for persistent storage")
	}
}

func TestDockerCompose_APIPort(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	if !strings.Contains(content, "8080") {
		t.Error("docker-compose.yml api service should expose port 8080")
	}
}

func TestDockerCompose_APIUsesEnvFile(t *testing.T) {
	root := projectRoot(t)
	content := readFile(t, filepath.Join(root, "docker-compose.yml"))

	if !strings.Contains(content, "env_file") {
		t.Error("docker-compose.yml api service should use env_file for configuration")
	}
}
