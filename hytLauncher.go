package main

import (
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"strconv"
	"os"
	"fmt"
	"path/filepath"
)

var HOME_FOLDER, _ = os.UserHomeDir();
var MAIN_FOLDER = filepath.Join(HOME_FOLDER, "hytLauncher");
var GAME_FOLDER = filepath.Join(MAIN_FOLDER, "game", "versions");
var USERDATA_FOLDER = filepath.Join(MAIN_FOLDER, "userdata");
var LAUNCHER_FOLDER = filepath.Join(MAIN_FOLDER, "launcher");
var JRE_FOLDER = filepath.Join(MAIN_FOLDER, "jre");

func urlToPath(targetUrl string) string {
	nurl, _ := url.Parse(targetUrl);
	npath := strings.TrimPrefix(nurl.Path, "/");
	return npath;
}

func download(targetUrl string, saveFilename string, progress func(done int64, total int64)) any {
	fmt.Printf("Downloading %s\n", targetUrl);

	os.MkdirAll(filepath.Dir(saveFilename), 0775);
	resp, err := http.Get(targetUrl);
	if err != nil {
		return nil;
	}

	if resp.StatusCode == 200 {
		f, _ := os.Create(saveFilename);

		defer f.Close();

		total := resp.ContentLength;
		done := int64(0);
		buffer := make([]byte, 0x8000);

		for done < total {
			rd, _ := resp.Body.Read(buffer);
			done += int64(rd);
			f.Write(buffer[:rd]);
			progress(done, total);
		}
	}

	return saveFilename;
}


func getVersionDownloadPath(startVersion int, endVersion int, channel string) string {
	fp := filepath.Join(GAME_FOLDER, "download", channel, strconv.Itoa(endVersion), strconv.Itoa(startVersion) + "-" + strconv.Itoa(endVersion)+".pwr");
	return fp;
}

func getVersionInstallPath(endVersion int, channel string) string {
	fp := filepath.Join(GAME_FOLDER, channel, strconv.Itoa(endVersion));
	return fp;
}

func getJrePath(operatingSystem string, architecture string) string {
	fp := filepath.Join(JRE_FOLDER, operatingSystem, architecture);
	return fp;
}

func getJreDownloadPath(operatingSystem string, architecture string, downloadUrl string) string {
	u, _ := url.Parse(downloadUrl);
	fp := filepath.Join(JRE_FOLDER, "download", operatingSystem, architecture, path.Base(u.Path));
	return fp;
}


func downloadLatestVersion(atokens accessTokens, architecture string, operatingSystem string, channel string, fromVersion int, progress func(done int64, total int64)) any {
	fmt.Printf("Start version: %d\n", fromVersion);
	manifest := getVersionManifest(atokens, architecture, operatingSystem, channel, fromVersion);
	for _, step := range manifest.Steps {
		save := getVersionDownloadPath(step.From, step.To, channel);
		return download(step.Pwr, save, progress);
	}
	return nil;
}


func installJre(progress func(done int64, total int64)) any {
	jres, ok := getJres("release").(versionFeed);
	if ok {
		downloadUrl := jres.DownloadUrls.Windows.Amd64.URL;
		save := getJreDownloadPath(runtime.GOOS, runtime.GOARCH, downloadUrl);
		unpack := getJrePath(runtime.GOOS, runtime.GOARCH);

		_, err := os.Stat(unpack);

		if err != nil {
			_, ok := download(downloadUrl, save, progress).(string);
			if ok {
				os.MkdirAll(unpack, 0775);

				unzip(save, unpack);
				os.Remove(save);
				os.RemoveAll(filepath.Dir(save));
				return unpack;
			}
		}

	}

	return nil;
}

func isGameVersionInstalled(version int, channel string) bool {
	gameDir := getVersionInstallPath(version, channel);
	_, err := os.Stat(gameDir);
	if err != nil {
		return false;
	}
	return true;
}

func installGame(version int, channel string, progress func(done int64, total int64)) any {
	save := getVersionDownloadPath(0, version, channel);
	unpack := getVersionInstallPath(version, channel);

	if !isGameVersionInstalled(version, channel) {
		downloadUrl := guessPatchUrlNoAuth(runtime.GOARCH, runtime.GOOS, channel, 0, version);
		pwr := download(downloadUrl, save, progress);
		os.MkdirAll(unpack, 0775);

		if _, ok := pwr.(string); ok {
			applyPatch(unpack, unpack, save);

			os.Remove(save);
			os.RemoveAll(filepath.Dir(save));
			return unpack;
		}
	}
	return nil;
}

func findJavaBin() any {
	jrePath := getJrePath(runtime.GOOS, runtime.GOARCH);

	d, err := os.ReadDir(jrePath);
	if err != nil {
		fmt.Printf("err: %s\n", err);
		os.Exit(0);
	}

	for _, e := range d {
		if !e.IsDir() {
			continue;
		}

		if runtime.GOOS == "windows" {
			return filepath.Join(jrePath, e.Name(), "bin", "java.exe");
		} else {
			return filepath.Join(jrePath, e.Name(), "bin", "java");
		}
	}

	return nil;
}

func launchGame(version int, channel, username string, uuid string) {

	javaBin, _ := findJavaBin().(string);


	if runtime.GOOS == "windows" {
		appDir := getVersionInstallPath(version, channel)
		userDir := USERDATA_FOLDER
		hytaleClientBin := filepath.Join(appDir, "Client", "HytaleClient.exe");

		e := exec.Command(hytaleClientBin,
				"--app-dir",
				appDir,
				"--user-dir",
				userDir,
				"--java-exec",
				javaBin,
				"--auth-mode",
				"offline",
				"--uuid",
				uuid,
				"--name",
				username);

		fmt.Printf("Running: %s %s\n", hytaleClientBin, strings.Join(e.Args, " "))

		e.Start();

		//runServer(username, uuid);
	}
}
