// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package basedir

var dataDir = firstGetenv("", "APPDATA")
