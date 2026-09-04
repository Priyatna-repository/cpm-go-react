package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/config"
)

var (
	ErrFileTooLarge     = errors.New("file exceeds the maximum allowed size")
	ErrUnsupportedImage = errors.New("unsupported image type")
)

const maxLogoSize = 2 << 20 // 2MB — normalized to match the reference app's ClientCompany limit (its OwnerCompany form allowed 12MB inconsistently; we don't replicate that mismatch)

var allowedImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
}

type UploadService struct {
	cfg *config.Config
}

func NewUploadService(cfg *config.Config) *UploadService {
	return &UploadService{cfg: cfg}
}

// SaveImage validates and stores an uploaded image under uploads/<subdir>/,
// returning the public URL path to store on the owning record (e.g.
// "/uploads/client-companies/ab12cd34ef56.png").
func (s *UploadService) SaveImage(fh *multipart.FileHeader, subdir string) (string, error) {
	if fh.Size > maxLogoSize {
		return "", ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedImageExt[ext] {
		return "", ErrUnsupportedImage
	}

	dir := filepath.Join(s.cfg.UploadDir, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	name, err := randomFilename(ext)
	if err != nil {
		return "", err
	}

	if err := saveMultipartFile(fh, filepath.Join(dir, name)); err != nil {
		return "", err
	}

	return "/uploads/" + subdir + "/" + name, nil
}

// DeleteImage removes a previously-saved image given its public URL path
// (as returned by SaveImage). Missing files, or paths not under our
// upload dir, are silently ignored — not an error.
func (s *UploadService) DeleteImage(publicPath string) error {
	if publicPath == "" {
		return nil
	}
	rel := strings.TrimPrefix(publicPath, "/uploads/")
	if rel == publicPath {
		return nil // not one of ours, don't touch it
	}

	absUploadDir, err := filepath.Abs(s.cfg.UploadDir)
	if err != nil {
		return err
	}
	absFull, err := filepath.Abs(filepath.Join(s.cfg.UploadDir, rel))
	if err != nil {
		return err
	}
	// guard against a crafted path (e.g. "../../etc/passwd") escaping UploadDir
	if !strings.HasPrefix(absFull, absUploadDir+string(os.PathSeparator)) {
		return errors.New("invalid file path")
	}

	if err := os.Remove(absFull); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func randomFilename(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + ext, nil
}

func saveMultipartFile(fh *multipart.FileHeader, dest string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
