package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// fileInfo结构体存储文件路径和其MD5哈希值
type fileInfo struct {
	path string
	hash string
}

// 计算文件的MD5哈希值
func calculateHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// 处理单个文件，计算哈希值并返回fileInfo结构体
func processFile(path string, wg *sync.WaitGroup, mu *sync.Mutex, fileInfos map[string][]fileInfo) {
	defer wg.Done()

	hash, err := calculateHash(path)
	if err != nil {
		fmt.Printf("计算文件 %s 的哈希值时出错: %v\n", path, err)
		return
	}

	mu.Lock()
	fileInfos[hash] = append(fileInfos[hash], fileInfo{path: path})
	mu.Unlock()
}

// 查找指定目录下的重复文件
func findDuplicates(root string) map[string][]fileInfo {
	fileInfos := make(map[string][]fileInfo)
	var wg sync.WaitGroup
	var mu sync.Mutex

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			wg.Add(1)
			go processFile(path, &wg, &mu, fileInfos)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("遍历目录 %s 时出错: %v\n", root, err)
	}

	wg.Wait()

	return fileInfos
}

// 打印重复文件信息
func printDuplicates(duplicates map[string][]fileInfo) {
	for _, files := range duplicates {
		if len(files) > 1 {
			fmt.Println("重复文件组:")
			for _, file := range files {
				fmt.Printf("  %s\n", file.path)
			}
			fmt.Println()
		}
	}
}

func main() {
	root := "/Users/zen/Downloads/media"
	duplicates := findDuplicates(root)
	printDuplicates(duplicates)
}
