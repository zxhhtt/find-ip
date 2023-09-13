package main

import (
    "archive/zip"
    "bufio"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "os/exec"
)

func main() {
    // 获取当前工作目录
    currentDirectory, err := os.Getwd()
    if err != nil {
        log.Fatalf("获取当前工作目录失败：%v", err)
        return
    }

    // 删除当前文件夹中的所有txt和zip文件
    err = deleteFiles(currentDirectory, ".txt", ".zip")
    if err != nil {
        log.Fatalf("删除文件失败：%v", err)
        return
    }
    fmt.Println("清理旧文件 >>>")
	
    // 下载zip文件
    zipURL := "https://zip.baipiao.eu.org"
    zipFileName := "files.zip"
    err = downloadFile(zipURL, zipFileName)
    if err != nil {
        log.Fatalf("下载文件失败：%v", err)
        return
    }

    // 解压文件
    unzipDirectory := currentDirectory
    err = unzip(zipFileName, unzipDirectory)
    if err != nil {
        log.Fatalf("解压文件失败：%v", err)
        return
    }
    fmt.Println("下载并解压 >>>")
	
    // 合并txt文件并去重
    err = mergeAndRemoveDuplicates(unzipDirectory)
    if err != nil {
        log.Fatalf("合并和去重文件失败：%v", err)
        return
    }
    fmt.Println("合并与去重 >>>")

    // 复制到目标路径
    destinationDirectory := "D:\\Download\\ip\\0724workers-IP"
    err = copyFile(filepath.Join(unzipDirectory, "ip.txt"), filepath.Join(destinationDirectory, "ip.txt"))
    if err != nil {
        log.Fatalf("复制文件失败：%v", err)
        return
    }
    fmt.Println("复制新文件 >>>")

    // 删除包含"-0-"的txt文件
    err = deleteFilesWithPattern(unzipDirectory, "*-0-*.txt")
    if err != nil {
        log.Fatalf("删除文件失败：%v", err)
        return
    }
    fmt.Println("del  80 >>>")
    fmt.Println("已完成")

    // 询问用户是否跳转到指定目录
    var userInput string
    fmt.Print("是否跳转到 D:\\Download\\ip\\0724workers-IP？ (Y/N): ")
    // 读取用户输入
    fmt.Scanln(&userInput)

    // 如果用户未输入任何内容，则默认为 "Y"
    if userInput == "" || strings.ToLower(userInput) == "y" {
        // 执行跳转到目录的操作
        err := openExplorer(destinationDirectory) 
        if err != nil {
            log.Fatalf("打开目录失败：%v", err)
        }
    }
}

func downloadFile(url, fileName string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    file, err := os.Create(fileName)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.Copy(file, resp.Body)
    if err != nil {
        return err
    }

    return nil
}

func unzip(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        rc, err := f.Open()
        if err != nil {
            return err
        }

        path := filepath.Join(dest, f.Name)
        if f.FileInfo().IsDir() {
            os.MkdirAll(path, os.ModePerm)
        } else {
            os.MkdirAll(filepath.Dir(path), os.ModePerm)
            outFile, err := os.Create(path)
            if err != nil {
                return err
            }
            defer outFile.Close()

            _, err = io.Copy(outFile, rc)
            if err != nil {
                return err
            }
        }
        rc.Close()
    }

    return nil
}

func mergeAndRemoveDuplicates(directory string) error {
    mergedFileName := "ip.txt"
    mergedFilePath := filepath.Join(directory, mergedFileName)
    mergedFile, err := os.Create(mergedFilePath)
    if err != nil {
        return err
    }
    defer mergedFile.Close()

    files, err := ioutil.ReadDir(directory)
    if err != nil {
        return err
    }

    uniqueLines := make(map[string]struct{})
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".txt") {
            filePath := filepath.Join(directory, file.Name())
            content, err := ioutil.ReadFile(filePath)
            if err != nil {
                return err
            }
            lines := strings.Split(string(content), "\n")
            for _, line := range lines {
                if len(line) > 0 {
                    uniqueLines[line] = struct{}{}
                }
            }
        }
    }

    writer := bufio.NewWriter(mergedFile)
    for line := range uniqueLines {
        _, err := writer.WriteString(line + "\n")
        if err != nil {
            return err
        }
    }
    return writer.Flush()
}

func copyFile(src, dest string) error {
    sourceFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    destFile, err := os.Create(dest)
    if err != nil {
        return err
    }

    defer destFile.Close()

    _, err = io.Copy(destFile, sourceFile)
    if err != nil {
        return err
    }

    return nil
}

func deleteFiles(directory string, extensions ...string) error {
    files, err := ioutil.ReadDir(directory)
    if err != nil {
        return err
    }

    for _, file := range files {
        for _, ext := range extensions {
            if strings.HasSuffix(file.Name(), ext) {
                filePath := filepath.Join(directory, file.Name())
                err := os.Remove(filePath)
                if err != nil {
                    return err
                }
            }
        }
    }
    return nil
}

func deleteFilesWithPattern(directory, pattern string) error {
    files, err := ioutil.ReadDir(directory)
    if err != nil {
        return err
    }

    for _, file := range files {
        if matched, _ := filepath.Match(pattern, file.Name()); matched {
            filePath := filepath.Join(directory, file.Name())
            err := os.Remove(filePath)
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func openExplorer(directory string) error {
    return exec.Command("explorer", directory).Start()
}
