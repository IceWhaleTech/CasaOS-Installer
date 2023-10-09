package internal_test

import (
	"testing"
)

// 回头补一下这个单元测试
func TestDecompressTar(t *testing.T) {
	// // create a tar file
	// // create a temp dir
	// // decompress tar file to temp dir

	// buf := new(bytes.Buffer)

	// // 创建一个 tar 写入器
	// tw := tar.NewWriter(buf)

	// // 添加一些文件头
	// // 通常需要设置名称、模拟文件内容、模式等
	// header := &tar.Header{
	// 	Name: "test.txt",
	// 	Mode: 0600,
	// 	Size: int64(len("Hello World")),
	// }
	// tw.WriteHeader(header)
	// tw.Write([]byte("Hello World"))
	// tw.Close()

	// // 用 tar 解压到临时目录
	// tmpDir := "testDIR"
	// defer os.RemoveAll(tmpDir)

	// err := internal.UnTar(buf.String(), tmpDir)
	// assert.NoError(t, err)

	// // 检查文件是否正确解压
	// data, err := os.ReadFile(tmpDir + "/test.txt")
	// assert.NoError(t, err)

	// assert.Equal(t, "Hello World", string(data))
}
