package launcher

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

type LauncherInterface interface {
	PrivilageCheck() error
	CheckPlaform() (architecture string, platform string)
	WritePubKey() error
	CosignCheck() error
	CosignInstall(architecture string, platform string) error
	ToolsInstall() error
}
type Linux struct {
}

func (*Linux) PrivilageCheck() error {
	currentUser, err := user.Current()
	if err != nil {
		println("Error getting current user:", err)
		return err
	}
	sudoUser := os.Getenv("SUDO_USER")
	if currentUser.Uid == "0" {
		if sudoUser != "" {
			fmt.Printf("This application was started with sudo by user %s. Exiting.\n", sudoUser)
			return nil
		} else {
			err = fmt.Errorf("This application should not be run as root. Exiting.")
			return err
		}
	}
	err = fmt.Errorf("Non-root user detected. Proceeding with the application.")
	return err
}

const (
	KEYS_DIR           = "/usr/keys"
	KIRA_COSIGN_PUB    = "/usr/keys/kira-cosign.pub"
	BASE_IMAGE_VERSION = "v0.13.5"
	TOOLS_VERSION      = "v0.3.42"
	COSIGN_VERSION     = "v2.0.0"
	COSIGN_HASH_ARM    = "8132cb2fb99a4c60ba8e03b079e12462c27073028a5d08c07ecda67284e0c88d"
	COSIGN_HASH_AMD    = "169a53594c437d53ffc401b911b7e70d453f5a2c1f96eb2a736f34f6356c4f2b"
	FILE_NAME          = "bash-utils.sh"
)

func (*Linux) CosignCheck() error {
	_, err := exec.LookPath("cosign")
	if err != nil {
		return err
	} else {
		fmt.Println("cosign is installed")
		return nil
	}
}

// T O   D O
// I M P O R T A N T
// CHANGE IN FUTURE INTO sigstore/cosign -
// это библиотека на Go, предоставляющая API для работы с подписями и верификации контейнеров.
func (*Linux) CosignInstall(architecture string, platform string) error {
	fmt.Println("INSTALLING....")
	// wget 				  https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/$FILE_NAME && chmod +x -v ./$FILE_NAME
	cosignURL := fmt.Sprintf("https://github.com/sigstore/cosign/releases/download/%s/cosign-%s-%s", COSIGN_VERSION, platform, architecture)
	cosignFile := filepath.Base(cosignURL)
	fmt.Printf("Downloading cosign from %s\n", cosignURL)
	// resp, err := http.Get(cosignURL)
	// if err != nil {
	// 	fmt.Printf("Failed to download cosign: %v\n", err)
	// 	return err
	// }
	// defer resp.Body.Close()
	// out, err := os.Create(cosignFile)
	// if err != nil {
	// 	fmt.Printf("Failed to download cosign: %v\n", err)
	// 	return err
	// }
	// defer out.Close()
	// if _, err := io.Copy(out, resp.Body); err != nil {
	// 	fmt.Printf("Failed to copy cosign: %v\n", err)
	// 	return err
	// }
	downloadFile(cosignURL, cosignFile)

	hashCmd := exec.Command("sha256sum", cosignFile)
	hashOut, err := hashCmd.Output()
	if err != nil {
		fmt.Printf("Failed to compute sha256sum: %v\n", err)
		return err
	}

	hash := strings.Fields(string(hashOut))[0]
	switch platform {
	case "arm64":
		if hash != COSIGN_HASH_ARM {
			err = fmt.Errorf("Invalid checksum for cosign: %s\n", hash)
			return err
		}
	case "amd64":
		if hash != COSIGN_HASH_AMD {
			err = fmt.Errorf("Invalid checksum for cosign: %s\n", hash)
			return err
		}
	}

	if err := os.Chmod(cosignFile, 0755); err != nil {
		err = fmt.Errorf("Failed to make cosign executable: %v\n", err)
		return err
	}
	if err := os.Rename(cosignFile, "/usr/local/bin/cosign"); err != nil {
		err = fmt.Errorf("Failed to install cosign: %v\n", err)
		return err
	}

	// Create keys directory if it doesn't exist
	err = os.MkdirAll(KEYS_DIR, 0755)
	if err != nil {
		fmt.Printf("Error creating keys directory: %v\n", err)
		return err
	}

	return nil
}
func (*Linux) WritePubKey() error {
	// Write public key to file
	pubKey := `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/IrzBQYeMwvKa44/DF/HB7XDpnE+
f+mU9F/Qbfq25bBWV2+NlYMJv3KvKHNtu3Jknt6yizZjUV4b8WGfKBzFYw==
-----END PUBLIC KEY-----`
	f, err := os.Create(KIRA_COSIGN_PUB)
	if err != nil {
		err = fmt.Errorf("Error creating file: %v\n", err)
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s", pubKey)
	if err != nil {
		err = fmt.Errorf("Error writing to file: %v\n", err)
		return err
	}
	fmt.Println("Public key written to file")
	return nil
}
func (*Linux) CheckPlaform() (architecture string, platform string) {
	architecture = runtime.GOARCH
	platform = runtime.GOOS
	if strings.Contains(architecture, "arm") {
		architecture = "arm64"
	} else {
		architecture = "amd64"
	}
	fmt.Printf("ARCH: %s, PLATFORM: %s\n", architecture, platform)
	return architecture, platform
}

//KIRA TOOLS INSTALATION
//

func (*Linux) ToolsInstall() error {
	// Download the tools binary and its signature
	toolsURL := fmt.Sprintf("https://github.com/KiraCore/tools/releases/download/%s/%s", TOOLS_VERSION, FILE_NAME)
	sigURL := fmt.Sprintf("https://github.com/KiraCore/tools/releases/download/%s/%s.sig", TOOLS_VERSION, FILE_NAME)
	fmt.Printf("Downloading %s and its signature...\n", FILE_NAME)
	if err := downloadFile(toolsURL, FILE_NAME); err != nil {
		fmt.Printf("Failed to download %s: %v\n", FILE_NAME, err)
		return err
	}
	if err := downloadFile(sigURL, FILE_NAME+".sig"); err != nil {
		fmt.Printf("Failed to download %s signature: %v\n", FILE_NAME, err)
		return err
	}
	// Verify the signature using cosign
	fmt.Println("Verifying the signature...")
	if err := verifySignature(FILE_NAME, FILE_NAME+".sig"); err != nil {
		fmt.Printf("Failed to verify signature: %v\n", err)
		return err
	}
	// Make the tools binary executable
	if err := os.Chmod(FILE_NAME, 0755); err != nil {
		fmt.Printf("Failed to make %s executable: %v\n", FILE_NAME, err)
		return err
	}
	// Run the tools binary to install bash-utils
	fmt.Println("Installing bash-utils...")
	if err := runCommand("./"+FILE_NAME, "bashUtilsSetup", "/var/kiraglob"); err != nil {
		fmt.Printf("Failed to install bash-utils: %v\n", err)
		return err
	}
	// Reload the profile
	if err := runCommand(". /etc/profile"); err != nil {
		fmt.Printf("Failed to reload profile: %v\n", err)
		return err
	}
	fmt.Printf("Installed bash-utils %s\n", bashUtilsVersion())
	return nil
}
func downloadFile(url, fileName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

func verifySignature(fileName, sigName string) error {
	cosignCmd := exec.Command("cosign", "verify-blob", "--key", os.Getenv("KIRA_COSIGN_PUB"), "--signature", sigName, fileName)
	if err := cosignCmd.Run(); err != nil {
		return err
	}
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}