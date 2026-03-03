package scanner

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// binarySniffSize はバイナリ判定で読み取る先頭バイト数。
// Git と同等の 8KB を採用（git の diff.c: FIRST_FEW_BYTES = 8000）。
const binarySniffSize = 8192

// binaryExtensions は既知のバイナリファイル拡張子のセット。
// 拡張子チェックで高速にスキップするために使用する。
var binaryExtensions = map[string]bool{
	// 画像
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".bmp": true, ".ico": true, ".webp": true, ".tiff": true,
	".tif": true, ".avif": true, ".heic": true, ".heif": true,

	// 音声
	".mp3": true, ".wav": true, ".ogg": true, ".flac": true,
	".aac": true, ".wma": true, ".m4a": true,

	// 動画
	".mp4": true, ".avi": true, ".mkv": true, ".mov": true,
	".wmv": true, ".flv": true, ".webm": true, ".m4v": true,

	// アーカイブ
	".zip": true, ".tar": true, ".gz": true, ".bz2": true,
	".xz": true, ".7z": true, ".rar": true, ".zst": true,
	".jar": true, ".war": true, ".ear": true,

	// 実行ファイル・ライブラリ
	".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".bin": true, ".o": true, ".a": true, ".lib": true,
	".class": true, ".wasm": true,

	// ドキュメント（バイナリ形式）
	".pdf": true, ".doc": true, ".docx": true,
	".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,

	// フォント
	".woff": true, ".woff2": true, ".ttf": true, ".otf": true,
	".eot": true,

	// データベース
	".db": true, ".sqlite": true, ".sqlite3": true,
}

// isBinary はファイルがバイナリかどうかを2段階で判定する。
// 第1段階: 拡張子チェック（I/O なし）
// 第2段階: ファイル先頭の null バイト検出
func isBinary(path string) (bool, error) {
	if isBinaryExtension(path) {
		return true, nil
	}
	return isBinaryContent(path)
}

// isBinaryExtension は拡張子が既知のバイナリ形式かを判定する。
func isBinaryExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return binaryExtensions[ext]
}

// isBinaryContent はファイル先頭を読み取り、null バイトの有無でバイナリ判定する。
func isBinaryContent(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, binarySniffSize)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	if n == 0 {
		return false, nil
	}

	return bytes.Contains(buf[:n], []byte{0x00}), nil
}
