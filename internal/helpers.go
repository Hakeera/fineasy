package storage

import "path/filepath"

// dirOf retorna o diretório de um caminho de arquivo.
// Usado para garantir que os diretórios existam antes de criar os CSVs.
func DirOf(filePath string) string {
	return filepath.Dir(filePath)
}
