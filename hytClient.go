package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

const ACCOUNT_DATA_URL = "https://account-data.hytale.com/";
const GAME_PATCHES_URL = "https://game-patches.hytale.com/";
const LAUNCHER_URL     = "https://launcher.hytale.com/"

func guessPatchSigUrlNoAuth(architecture string, operatingSystem string, channel string, startVersion int, targetVersion int) string{
	fullUrl, _ := url.JoinPath(GAME_PATCHES_URL, "patches", operatingSystem, architecture, channel, strconv.Itoa(startVersion), strconv.Itoa(targetVersion) + ".pwr.sig");
	return fullUrl;
}
func guessPatchUrlNoAuth(architecture string, operatingSystem string, channel string, startVersion int, targetVersion int) string{
	fullUrl, _ := url.JoinPath(GAME_PATCHES_URL, "patches", operatingSystem, architecture, channel, strconv.Itoa(startVersion), strconv.Itoa(targetVersion) + ".pwr");
	return fullUrl;
}

func getJres(channel string) versionFeed {
	fullUrl, _ := url.JoinPath(LAUNCHER_URL, "version", channel, "jre.json");

	resp, _ := http.Get(fullUrl);

	feed := versionFeed{};
	json.NewDecoder(resp.Body).Decode(&feed);

	return feed;
}

func getLaunchers(channel string) versionFeed {
	fullUrl, _ := url.JoinPath(LAUNCHER_URL, "version", channel, "launcher.json");

	resp, _ := http.Get(fullUrl);

	feed := versionFeed{};
	json.NewDecoder(resp.Body).Decode(&feed);

	return feed;

}

func getLauncherData(atokens accessTokens, architecture string, operatingSystem string) launcherData{

	fullUrl, _ := url.JoinPath(ACCOUNT_DATA_URL, "my-account", "get-launcher-data");
	launcherDataUrl, _ := url.Parse(fullUrl);

	query := make(url.Values);
	query.Add("arch", architecture);
	query.Add("os", operatingSystem);

	launcherDataUrl.RawQuery = query.Encode();

	req, _:= http.NewRequest("GET", launcherDataUrl.String(), nil);

	req.Header.Add("Authorization", "Bearer " + atokens.AccessToken);
	req.Header.Add("Content-Type", "application/json");

	resp, _ := http.DefaultClient.Do(req);

	ldata := launcherData{};
	json.NewDecoder(resp.Body).Decode(&ldata);

	return ldata;
}

func getVersionManifest(atokens accessTokens, architecture string, operatingSystem string, channel string, gameVersion int) versionManifest {
	fullUrl, _ := url.JoinPath(ACCOUNT_DATA_URL, "patches", operatingSystem, architecture, channel, strconv.Itoa(gameVersion));

	req, _:= http.NewRequest("GET", fullUrl, nil);

	req.Header.Add("Authorization", "Bearer " + atokens.AccessToken);
	req.Header.Add("Content-Type", "application/json");

	resp, _ := http.DefaultClient.Do(req);

	mdata := versionManifest{};
	json.NewDecoder(resp.Body).Decode(&mdata);

	return mdata;
}
