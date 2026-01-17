package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"bitbucket.org/rj/goey"
	"bitbucket.org/rj/goey/base"
	"bitbucket.org/rj/goey/loop"
	"bitbucket.org/rj/goey/windows"
)

var LAUNCHER_JSON = filepath.Join(LAUNCHER_FOLDER, "launcher.json");

type launcherState struct {
	Patchline string `json:"last_patchline"`
	Username string `json:"last_username"`
	SelectedVersion int `json:"last_version"`
	LatestVersions map[string]int `json:"last_version_scan_result"`
}


var(
	wMainWin *windows.Window
	wCommune launcherState;
	wProgress = 0
	wDisabled = false
)



func checkForUpdates() {
	lastRelease := wCommune.LatestVersions["release"]
	lastPreRelease := wCommune.LatestVersions["pre-release"]

	latestRelease := findLatestVersionNoAuth(lastRelease, runtime.GOARCH, runtime.GOOS, "release");
	latestPreRelease := findLatestVersionNoAuth(lastPreRelease, runtime.GOARCH, runtime.GOOS, "pre-release");

	fmt.Printf("latestRelease: %d\n", latestRelease);
	fmt.Printf("latestPreRelease: %d\n", latestPreRelease);

	if latestRelease > lastRelease {
		fmt.Printf("Found new release version: %d", latestRelease);
		wCommune.LatestVersions["release"] = latestRelease;
	}

	if latestPreRelease > lastPreRelease {
		fmt.Printf("Found new pre-release version: %d", latestPreRelease);
		wCommune.LatestVersions["pre-release"] = latestPreRelease;
	}

	if wMainWin != nil {
		updateWindow();
		writeSettings();
	}
}

func writeSettings() {
	fmt.Printf("Saving settings ...\n");
	jlauncher, _ := json.Marshal(wCommune);

	err := os.MkdirAll(filepath.Dir(LAUNCHER_JSON), 0666);
	if err != nil {
		fmt.Printf("error writing settings: %s\n", err);
		return;
	}


	err = os.WriteFile(LAUNCHER_JSON, jlauncher, 0666);
	if err != nil {
		fmt.Printf("error writing settings: %s\n", err);
		return;
	}
}

func getDefaultSettings() {
	wCommune.Patchline = "release";
	wCommune.LatestVersions = map[string]int{
		"release": 4,
		"pre-release": 8,
	};
	wCommune.SelectedVersion = wCommune.LatestVersions[wCommune.Patchline];
	wCommune.Username = "TransRights";
	writeSettings();

	// check for updates in the background
	go checkForUpdates();

}

func readSettings() {
	_, err := os.Stat(LAUNCHER_JSON)
	if err != nil {
		getDefaultSettings();
	} else {
		data, err := os.ReadFile(LAUNCHER_JSON);
		if err != nil{
			getDefaultSettings();
			return;
		}
		json.Unmarshal(data, &wCommune);

		fmt.Printf("Reading last settings: \n");
		fmt.Printf("username: %s\n", wCommune.Username);
		fmt.Printf("patchline: %s\n", wCommune.Patchline);
		fmt.Printf("last used version: %d\n", wCommune.SelectedVersion);
		fmt.Printf("newest known release: %d\n", wCommune.LatestVersions["release"])
		fmt.Printf("newest known pre-release: %d\n", wCommune.LatestVersions["pre-release"])

		// check for updates in the background
		go checkForUpdates();
	}
}


func valToChannel(vchl int) string {
	switch vchl {
		case 0:
			return "release";
		case 1:
			return "pre-release";
		default:
			return "release";
	}
}

func channelToVal(channel string) int {
	switch channel {
		case "release":
			return 0;
		case "pre-release":
			return 1;
		default:
			return 0;
	}
}

func patchLineMenu() base.Widget {
	return &goey.VBox{
		AlignMain: goey.SpaceBetween,
		Children: []base.Widget{
			&goey.Label{Text: "Patchline:"},
			&goey.SelectInput{
				Items: []string {
					"release",
					"pre-release",
				},
				Value: channelToVal(wCommune.Patchline),
				Disabled: wDisabled,
				OnChange: func(v int) {
					if wCommune.Patchline != valToChannel(v) {
						wCommune.Patchline = valToChannel(v);
						updateWindow();
					}
				},
			},
		},
	};
}


func versionMenu() base.Widget {
	versions := goey.SelectInput {
		OnChange: func(v int) { wCommune.SelectedVersion = v+1},
		Disabled: wDisabled,
	};
	latest := wCommune.LatestVersions[wCommune.Patchline];

	for i := range latest {
		txt := "Version "+strconv.Itoa(i+1);
		if isGameVersionInstalled(i+1, wCommune.Patchline) {
			txt += " - installed";
		} else {
			txt += " - not installed";
		}

		versions.Items = append(versions.Items, txt);
	}

	versions.Value = wCommune.SelectedVersion;

	return &goey.VBox{
		AlignMain: goey.SpaceBetween,
		Children: []base.Widget{
			&goey.Label{Text: "Version:"},
			&versions,
		},
	};
}

func usernameBox() base.Widget {
	return &goey.VBox{
		AlignMain: goey.SpaceBetween,
		Children: []base.Widget{
			&goey.Label{Text: "Username:"},
			&goey.TextInput{
					Value: wCommune.Username,
					Placeholder: "Username",
					Disabled: wDisabled,
					OnChange: func(v string) {
						wCommune.Username = v;
					},
			},
		},
	};
}

func updateProgress(done int64, total int64) {
	lastProgress := wProgress;
	wProgress = int((float64(done) / float64(total)) * 100.0);
	if lastProgress != wProgress{
		updateWindow();
	}
}

func createWindow() error {
	w, err := windows.NewWindow("hytLauncher", renderWindow())
	if err != nil {
		return err
	}

	w.SetScroll(false, false);

	w.SetOnClosing(func() bool {
			writeSettings();
			return false;
	});

	wMainWin = w;


	return nil
}


func updateWindow() {
	err := wMainWin.SetChild(renderWindow())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}

func renderWindow() base.Widget {
	return &goey.Padding{
		Insets: goey.DefaultInsets(),
		Child: &goey.Align{
			Child: &goey.VBox{
				AlignMain: goey.MainStart,
				Children: []base.Widget{

					usernameBox(),
					patchLineMenu(),
					versionMenu(),

					&goey.Progress{
						Value: wProgress,
						Min: 0,
						Max: 100,
					},
					&goey.Button{
						Text: "Start Game",
						Disabled: wDisabled,
						OnClick: func() {
							go func() {
								wDisabled = true;
								installJre(updateProgress);
								installGame(wCommune.SelectedVersion, wCommune.Patchline, updateProgress);
								launchGame(wCommune.SelectedVersion, wCommune.Patchline, wCommune.Username, usernameToUuid(wCommune.Username));
								wDisabled = false;

								updateWindow();
							}();
						},
					},
				},
			},
		},
	}
}


func main() {

	os.MkdirAll(MAIN_FOLDER, 0666);
	os.MkdirAll(GAME_FOLDER, 0666);
	os.MkdirAll(USERDATA_FOLDER, 0666);
	os.MkdirAll(JRE_FOLDER, 0666);

	readSettings();

	err := loop.Run(createWindow)
	if err != nil {
		fmt.Println("Error: ", err)
	}

}
