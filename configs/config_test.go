package configs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitCreatesMissingConfigInUserDirectory(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := Init(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if !viper.GetBool(GameAutoConfirm) {
		t.Fatal("默认配置未初始化")
	}
}

func TestInitNeverOverwritesInvalidConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "config.yaml")
	original := []byte("game: [broken\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := Init(path); err == nil {
		t.Fatal("无效配置必须返回错误")
	}
	actual, err := os.ReadFile(path)
	if err != nil || string(actual) != string(original) {
		t.Fatal("原配置被覆盖")
	}
}

func TestInitReportsUnusablePathWithoutPanic(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parent, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Init(filepath.Join(parent, "config.yaml")); err == nil {
		t.Fatal("路径错误必须返回")
	}
}
