// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package basedir

var (
	cacheDir = firstGetenv("", "TEMP", "TMP")
	dataDir  = firstGetenv("", "APPDATA")
)

func cache() string {
	return cacheDir
}

func data() string {
	return dataDir
}
