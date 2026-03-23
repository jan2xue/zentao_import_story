// Package main 是发布打包工具
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	projectName = "zentao_import_story"
	defaultVersion = "1.1.0"
)

func main() {
	version := flag.String("version", defaultVersion, "发布版本号")
	outputDir := flag.String("output", "release", "输出目录")
	flag.Parse()

	fmt.Println("========================================")
	fmt.Println("  Zentao Story Importer - Release Build")
	fmt.Printf("  Version: v%s\n", *version)
	fmt.Println("========================================")
	fmt.Println()

	distDir := filepath.Join(*outputDir, fmt.Sprintf("%s_v%s", projectName, *version))

	// 1. 清理旧的发布目录
	fmt.Println("[1/6] Cleaning old release directory...")
	if _, err := os.Stat(*outputDir); err == nil {
		if err := os.RemoveAll(*outputDir); err != nil {
			fmt.Printf("清理失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 2. 创建发布目录
	fmt.Println("[2/6] Creating release directory...")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		os.Exit(1)
	}

	// 3. 编译可执行文件
	fmt.Println("[3/6] Building executable...")
	exeName := "zentao_story_tool"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}
	exePath := filepath.Join(distDir, exeName)

	// 设置编译参数
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", exePath, "./cmd/zentao_tool")
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS=windows",
		"GOARCH=amd64",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("编译失败: %v\n", err)
		os.Exit(1)
	}

	// 4. 复制配置文件
	fmt.Println("[4/6] Copying config file...")
	if err := copyFile("config.example.yaml", filepath.Join(distDir, "config.yaml")); err != nil {
		fmt.Printf("复制配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 5. 复制文档和示例文件
	fmt.Println("[5/6] Copying documents and sample files...")
	filesToCopy := []string{
		"requirements.xlsx",
		"changelog.txt",
		"使用说明.txt",
	}
	for _, f := range filesToCopy {
		if _, err := os.Stat(f); err == nil {
			if err := copyFile(f, filepath.Join(distDir, f)); err != nil {
				fmt.Printf("复制 %s 失败: %v\n", f, err)
			}
		}
	}

	// 6. 创建 ZIP 压缩包
	fmt.Println("[6/6] Creating zip archive...")
	zipFile := filepath.Join(*outputDir, fmt.Sprintf("%s_v%s.zip", projectName, *version))
	if err := createZip(distDir, zipFile); err != nil {
		fmt.Printf("创建压缩包失败: %v\n", err)
		os.Exit(1)
	}

	// 显示结果
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Build completed!")
	fmt.Println("========================================")
	fmt.Println()

	// 列出发布目录内容
	fmt.Println("Release contents:")
	filepath.Walk(distDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(distDir, path)
			fmt.Printf("  - %s\n", relPath)
		}
		return nil
	})

	// 显示输出文件信息
	fmt.Println()
	fmt.Println("Output file:")
	fmt.Printf("  %s\n", zipFile)
	if info, err := os.Stat(zipFile); err == nil {
		fmt.Printf("  Size: %.2f KB\n", float64(info.Size())/1024)
	}
	fmt.Println()
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// createZip 创建 ZIP 压缩包
func createZip(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建 ZIP 条目
		relPath, err := filepath.Rel(filepath.Dir(sourceDir), path)
		if err != nil {
			return err
		}

		// Windows 路径分隔符转换为 ZIP 标准的斜杠
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		if info.IsDir() {
			relPath += "/"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		if info.IsDir() {
			header.Name += "/"
			_, err = zipWriter.CreateHeader(header)
			return err
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}
