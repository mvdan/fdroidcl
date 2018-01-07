// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

// +build dragonfly freebsd linux netbsd openbsd

package basedir

var (
	cacheDir = firstGetenv(".cache", "XDG_CACHE_HOME")
	dataDir  = firstGetenv(".config", "XDG_CONFIG_HOME")
)

func cache() string {
	return cacheDir
}

func data() string {
	return dataDir
}
