// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

// +build dragonfly freebsd linux netbsd openbsd

package basedir

var dataDir = firstGetenv(".config", "XDG_CONFIG_HOME")
