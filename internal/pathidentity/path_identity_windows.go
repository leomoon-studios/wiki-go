package pathidentity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

var getFinalPathNameByHandle = syscall.NewLazyDLL("kernel32.dll").NewProc("GetFinalPathNameByHandleW")

func validateIdentityPath(value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("Windows path is not valid UTF-8")
	}
	for _, component := range strings.Split(value, "/") {
		if component == "" || component == "." {
			continue
		}
		if strings.HasSuffix(component, ".") || strings.HasSuffix(component, " ") || !filepath.IsLocal(component) {
			return fmt.Errorf("invalid Windows path component %q", component)
		}
		for _, character := range component {
			if character < 32 || strings.ContainsRune(`:<>"|?*`, character) {
				return fmt.Errorf("unsupported Windows path character in %q", component)
			}
		}
	}
	return nil
}

func existingPathIdentity(filename string) (string, error) {
	// Open the resource itself, following symlinks and junctions. EvalSymlinks
	// does not follow junctions with Go's default Windows reparse-point modes.
	// Use the extended namespace internally for long absolute filesystem paths;
	// URL-derived namespace forms have already been rejected before lookup.
	openPath := filename
	if !strings.HasPrefix(openPath, `\\?\`) {
		if strings.HasPrefix(openPath, `\\`) {
			openPath = `\\?\UNC\` + openPath[2:]
		} else {
			openPath = `\\?\` + openPath
		}
	}
	input, err := syscall.UTF16PtrFromString(openPath)
	if err != nil {
		return "", err
	}
	handle, err := syscall.CreateFile(input, 0, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", &os.PathError{Op: "open path identity", Path: filename, Err: err}
	}
	defer syscall.CloseHandle(handle)

	buffer := make([]uint16, 256)
	for {
		// Flags zero request FILE_NAME_NORMALIZED | VOLUME_NAME_DOS: the
		// resolved long names and stored case of the actual opened resource.
		n, _, callErr := getFinalPathNameByHandle.Call(uintptr(handle), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0)
		if n == 0 {
			return "", &os.PathError{Op: "query path identity", Path: filename, Err: callErr}
		}
		if n >= uintptr(len(buffer)) {
			if n > 32768 {
				return "", fmt.Errorf("Windows path identity exceeds the supported path length")
			}
			buffer = make([]uint16, n+1)
			continue
		}
		resolved := syscall.UTF16ToString(buffer[:n])
		switch {
		case strings.HasPrefix(resolved, `\\?\UNC\`):
			resolved = `\\` + resolved[len(`\\?\UNC\`):]
		case strings.HasPrefix(resolved, `\\?\`) && len(resolved) >= 7 && resolved[5] == ':' && resolved[6] == '\\':
			resolved = resolved[len(`\\?\`):]
		default:
			return "", fmt.Errorf("unsupported Windows path identity namespace")
		}
		return filepath.Clean(resolved), nil
	}
}
