package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	repo "lesiw.io/repo/lib"
)

const syncScript = `set -ex
builtin fc -R -I ~/.zsh_history
builtin fc -R -I %[1]s
builtin fc -W ~/.zsh_history
builtin fc -W %[1]s
`

func sync() error {
	dir, err := repoDir(os.Getenv("ZYNCREPO"))
	if err != nil {
		return err
	}
	f, err := os.Open(filepath.Join(dir, ".zsh_history"))
	if err != nil {
		return fmt.Errorf("could not open synced .zsh_history: %w", err)
	}
	defer func() { _ = f.Close() }()
	scratch, err := os.CreateTemp(dir, "")
	if err != nil {
		return fmt.Errorf("could not create scratch file: %w", err)
	}
	defer func() { _ = os.Remove(scratch.Name()) }()
	var clear string
	if fi, err := f.Stat(); err == nil && fi.Size() > 0 {
		clear, err = decrypt(f, os.Getenv("ZYNCPASS"))
		if err != nil {
			return fmt.Errorf("could not decrypt history: %w", err)
		}
	}
	if _, err := io.Copy(scratch, strings.NewReader(clear)); err != nil {
		return fmt.Errorf("could not copy decrypted history: %w", err)
	}
	buf, err := exec.Command("/bin/zsh", "-ci",
		fmt.Sprintf(syncScript, scratch.Name())).CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not sync history: %w\n%s", err, string(buf))
	}
	if buf, err = os.ReadFile(scratch.Name()); err != nil {
		return fmt.Errorf("could not read modified scratch file: %w", err)
	}
	if f, err = os.Create(filepath.Join(dir, ".zsh_history")); err != nil {
		return fmt.Errorf("could not reopen synced .zsh_history: %w", err)
	}
	if err = encrypt(f, string(buf), os.Getenv("ZYNCPASS")); err != nil {
		return fmt.Errorf("could not write encrypted history file: %w", err)
	}
	if err = pushFile(dir, ".zsh_history", "Update .zsh_history"); err != nil {
		return fmt.Errorf("could not push .zsh_history file: %w", err)
	}
	return nil
}

func repoDir(url string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache directory: %w", err)
	}
	dir, err := repo.Clone(filepath.Join(cache, "zync"), url, false)
	if err != nil {
		return "", fmt.Errorf("failed to fetch repository: %w", err)
	}
	if _, err := rnr.Get("git", "-C", dir, "pull", "--rebase"); err != nil {
		return "", fmt.Errorf("failed to fetch repository updates: %w", err)
	}

	return dir, nil
}

func pushFile(repo, path, msg string) error {
	_, err := rnr.Get("git", "-C", repo, "add", filepath.ToSlash(path))
	if err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	r, err := rnr.Get("git", "-C", repo, "status", "-s")
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}
	if r.Out == "" {
		return nil // Nothing changed, nothing to commit.
	}
	_, err = rnr.Get("git", "-C", repo, "commit", "-am", msg)
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	_, err = rnr.Get("git", "-C", repo, "push")
	if err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}
	return nil
}

const (
	keysz  = 32
	saltsz = 16
)

func key(pass string, salt []byte) []byte {
	return pbkdf2.Key([]byte(pass), salt, 4096, keysz, sha256.New)
}

func encrypt(w io.Writer, text, pass string) error {
	salt := make([]byte, saltsz)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	block, err := aes.NewCipher(key(pass, salt))
	if err != nil {
		return fmt.Errorf("failed to create cipher block: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	var ctext []byte
	ctext = append(ctext, salt...)
	ctext = append(ctext, nonce...)
	ctext = append(ctext, gcm.Seal(nil, nonce, []byte(text), nil)...)

	encoder := base64.NewEncoder(base64.StdEncoding, w)
	defer func() { _ = encoder.Close() }()
	if _, err = encoder.Write(ctext); err != nil {
		return err
	}

	return nil
}

func decrypt(r io.Reader, pass string) (string, error) {
	decoder := base64.NewDecoder(base64.StdEncoding, r)
	ctext, err := io.ReadAll(decoder)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	salt := ctext[:saltsz]
	ctext = ctext[saltsz:]

	block, err := aes.NewCipher(key(pass, salt))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create gcm: %w", err)
	}

	nonce := ctext[:gcm.NonceSize()]
	ctext = ctext[gcm.NonceSize():]

	text, err := gcm.Open(nil, nonce, ctext, nil)
	if err != nil {
		return "", err
	}

	return string(text), nil
}
