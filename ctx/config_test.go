package ctx

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const configChildFlag = "GO_CTX_CONFIG_CHILD"

func TestConfigurationCasingAndPrecedence(t *testing.T) {
	if os.Getenv(configChildFlag) == "1" {
		runConfigurationChild(t)
		return
	}
	command := exec.Command(os.Args[0], "-test.run=^TestConfigurationCasingAndPrecedence$")
	command.Env = append(os.Environ(), configChildFlag+"=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("configuration child failed: %v\n%s", err, output)
	}
}

func runConfigurationChild(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, ".env"), []byte("mixed_file=file\nfile_only=file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".env_custom"), []byte("file_only=custom\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(workingDirectory) }()

	SetEnv("default_only", "default")
	SetEnv("mixed_file", "default")
	os.Args = append(os.Args, "--mixed_file=argument")
	if err := os.Setenv("mixed_file", "exact-process"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("MIXED_FILE", "canonical-process"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("mixed_file")
	defer os.Unsetenv("MIXED_FILE")

	if got := GetEnv("mixed_file").AsString(); got != "exact-process" {
		t.Fatalf("exact process value = %q", got)
	}
	if err := os.Unsetenv("mixed_file"); err != nil {
		t.Fatal(err)
	}
	if got := GetEnv("mixed_file").AsString(); got != "canonical-process" {
		t.Fatalf("canonical process value = %q", got)
	}
	if err := os.Setenv("mixed_file", ""); err != nil {
		t.Fatal(err)
	}
	if value := GetEnv("mixed_file"); !value.IsPresent() || value.AsString() != "" {
		t.Fatalf("present-empty exact value = %#v", value)
	}
	if err := os.Unsetenv("mixed_file"); err != nil {
		t.Fatal(err)
	}
	if err := os.Unsetenv("MIXED_FILE"); err != nil {
		t.Fatal(err)
	}
	if got := GetEnv("mixed_file").AsString(); got != "argument" {
		t.Fatalf("argument value = %q", got)
	}
	if got := GetEnv("file_only").AsString(); got != "custom" {
		t.Fatalf("custom-file value = %q", got)
	}
	if got := GetEnv("default_only").AsString(); got != "default" {
		t.Fatalf("default value = %q", got)
	}
}
