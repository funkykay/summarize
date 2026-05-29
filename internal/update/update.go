package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/funkykay/summarize/internal/buildinfo"
)

const (
	AppName         = "summarize"
	APIBase         = "https://api.github.com"
	GitHubBase      = "https://github.com"
	DefaultRepoSlug = "funkykay/summarize"
)

type Error struct {
	Message string
}

func (e Error) Error() string {
	return e.Message
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func Binary(repoSlug string, stdout io.Writer) error {
	resolvedRepoSlug := DetectRepoSlug(repoSlug)
	if runtime.GOOS == "windows" {
		PrintManualUpdateInstructions(resolvedRepoSlug, stdout)
		return nil
	}

	release, err := FetchLatestRelease(resolvedRepoSlug, http.DefaultClient)
	if err != nil {
		return err
	}

	if release.TagName == "" {
		return Error{Message: "Latest release does not contain a tag_name."}
	}

	currentVersion := NormalizeVersion(buildinfo.Version)
	targetVersion := NormalizeVersion(release.TagName)

	newer, err := IsRemoteVersionNewer(currentVersion, targetVersion)
	if err != nil {
		return err
	}
	if !newer {
		fmt.Fprintf(stdout, "%s %s is already up to date.\n", AppName, currentVersion)
		return nil
	}

	assetName, err := DetectPlatformAssetName()
	if err != nil {
		return err
	}
	downloadURL, err := SelectAssetDownloadURL(release, assetName)
	if err != nil {
		return err
	}

	currentBinaryPath, err := DetectCurrentBinaryPath()
	if err != nil {
		return err
	}
	downloadPath := filepath.Join(filepath.Dir(currentBinaryPath), filepath.Base(currentBinaryPath)+".new")

	if err := DownloadBinary(downloadURL, downloadPath, http.DefaultClient); err != nil {
		return err
	}
	if err := ReplaceBinary(downloadPath, currentBinaryPath); err != nil {
		return Error{Message: fmt.Sprintf("Downloaded %s %s to %s, but could not replace %s: %s", AppName, targetVersion, downloadPath, currentBinaryPath, err)}
	}

	fmt.Fprintf(stdout, "Updated %s from %s to %s. Restart the command to use the new version.\n", AppName, currentVersion, targetVersion)
	return nil
}

func LatestReleaseURL(repoSlug string) string {
	return GitHubBase + "/" + repoSlug + "/releases/latest"
}

func PrintManualUpdateInstructions(repoSlug string, stdout io.Writer) {
	fmt.Fprintf(stdout, "Automatic updates are not supported on Windows because the running .exe cannot be replaced safely.\n")
	fmt.Fprintf(stdout, "Please download and install the latest %s release manually:\n", AppName)
	fmt.Fprintln(stdout, LatestReleaseURL(repoSlug))
}

func DetectRepoSlug(repoSlug string) string {
	if repoSlug != "" {
		return repoSlug
	}

	if environmentRepo := os.Getenv("GITHUB_REPOSITORY"); environmentRepo != "" {
		return environmentRepo
	}

	return DefaultRepoSlug
}

func FetchLatestRelease(repoSlug string, client HTTPClient) (Release, error) {
	request, err := http.NewRequest(http.MethodGet, APIBase+"/repos/"+repoSlug+"/releases/latest", nil)
	if err != nil {
		return Release{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", AppName)

	response, err := client.Do(request)
	if err != nil {
		return Release{}, Error{Message: fmt.Sprintf("Unable to fetch latest GitHub release for %s: %s.", repoSlug, err)}
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Release{}, Error{Message: fmt.Sprintf("Unable to fetch latest GitHub release for %s: HTTP %d.", repoSlug, response.StatusCode)}
	}

	var release Release
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&release); err != nil {
		return Release{}, Error{Message: "Latest GitHub release response is not valid JSON."}
	}

	return release, nil
}

func SelectAssetDownloadURL(release Release, assetName string) (string, error) {
	for _, asset := range release.Assets {
		if asset.Name != assetName {
			continue
		}

		if asset.BrowserDownloadURL != "" {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", Error{Message: fmt.Sprintf("Latest release does not contain asset '%s'.", assetName)}
}

func NormalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}

func IsRemoteVersionNewer(currentVersion string, targetVersion string) (bool, error) {
	current, err := ParseVersion(currentVersion)
	if err != nil {
		return false, err
	}
	target, err := ParseVersion(targetVersion)
	if err != nil {
		return false, err
	}

	maxLength := len(current)
	if len(target) > maxLength {
		maxLength = len(target)
	}

	for i := 0; i < maxLength; i++ {
		currentPart := 0
		if i < len(current) {
			currentPart = current[i]
		}
		targetPart := 0
		if i < len(target) {
			targetPart = target[i]
		}

		if currentPart < targetPart {
			return true, nil
		}
		if currentPart > targetPart {
			return false, nil
		}
	}

	return false, nil
}

func ParseVersion(version string) ([]int, error) {
	normalized := NormalizeVersion(version)
	releasePart := normalized
	if index := strings.IndexAny(releasePart, "-+"); index >= 0 {
		releasePart = releasePart[:index]
	}
	parts := strings.Split(releasePart, ".")
	if len(parts) == 0 || releasePart == "" {
		return nil, Error{Message: fmt.Sprintf("Unsupported version format: %s", version)}
	}

	parsed := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, Error{Message: fmt.Sprintf("Unsupported version format: %s", version)}
		}

		for _, char := range part {
			if char < '0' || char > '9' {
				return nil, Error{Message: fmt.Sprintf("Unsupported version format: %s", version)}
			}
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, Error{Message: fmt.Sprintf("Unsupported version format: %s", version)}
		}
		parsed = append(parsed, value)
	}

	return parsed, nil
}

func DetectPlatformAssetName() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	if osName == "linux" {
		if arch == "amd64" {
			return "summarize-linux-x64", nil
		}
		if arch == "arm64" {
			return "summarize-linux-arm64", nil
		}
	}

	if osName == "darwin" && arch == "arm64" {
		return "summarize-macos-arm64", nil
	}

	return "", Error{Message: fmt.Sprintf("No standalone release asset available for platform %s/%s.", platformSystemName(osName), runtime.GOARCH)}
}

func DetectCurrentBinaryPath() (string, error) {
	executablePath, err := os.Executable()
	if err == nil && executablePath != "" {
		return filepath.Abs(executablePath)
	}

	return "", Error{Message: "Unable to determine current binary path."}
}

func DownloadBinary(downloadURL string, downloadPath string, client HTTPClient) error {
	if err := os.MkdirAll(filepath.Dir(downloadPath), 0o755); err != nil {
		return Error{Message: err.Error()}
	}

	request, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return Error{Message: err.Error()}
	}
	request.Header.Set("User-Agent", AppName)

	response, err := client.Do(request)
	if err != nil {
		return Error{Message: fmt.Sprintf("Unable to download release asset: %s.", err)}
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Error{Message: fmt.Sprintf("Unable to download release asset: HTTP %d.", response.StatusCode)}
	}

	temporaryFile, err := os.CreateTemp(filepath.Dir(downloadPath), "."+AppName+".")
	if err != nil {
		return Error{Message: err.Error()}
	}
	temporaryPath := temporaryFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(temporaryPath)
		}
	}()

	_, copyErr := io.Copy(temporaryFile, response.Body)
	closeErr := temporaryFile.Close()
	if copyErr != nil {
		return Error{Message: copyErr.Error()}
	}
	if closeErr != nil {
		return Error{Message: closeErr.Error()}
	}

	info, err := os.Stat(temporaryPath)
	if err != nil {
		return Error{Message: err.Error()}
	}
	if err := os.Chmod(temporaryPath, info.Mode()|0o111); err != nil {
		return Error{Message: err.Error()}
	}
	if err := os.Rename(temporaryPath, downloadPath); err != nil {
		return Error{Message: err.Error()}
	}

	cleanup = false
	return nil
}

func ReplaceBinary(newBinaryPath string, currentBinaryPath string) error {
	if runtime.GOOS == "windows" {
		return Error{Message: "automatic updates are not supported on Windows"}
	}

	currentInfo, err := os.Stat(currentBinaryPath)
	if err != nil {
		return Error{Message: err.Error()}
	}
	if !currentInfo.Mode().IsRegular() {
		return Error{Message: fmt.Sprintf("current binary is not a regular file: %s", currentBinaryPath)}
	}
	if err := os.Chmod(newBinaryPath, currentInfo.Mode().Perm()|0o111); err != nil {
		return Error{Message: err.Error()}
	}
	if err := os.Rename(newBinaryPath, currentBinaryPath); err != nil {
		return Error{Message: err.Error()}
	}

	return nil
}

func platformSystemName(goos string) string {
	switch goos {
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows"
	default:
		if goos == "" {
			return goos
		}
		return strings.ToUpper(goos[:1]) + goos[1:]
	}
}

func IsUpdateError(err error) bool {
	var updateError Error
	return errors.As(err, &updateError)
}
