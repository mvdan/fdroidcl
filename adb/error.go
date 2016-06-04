// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package adb

import (
	"errors"
	"fmt"
)

var (
	// Common install and uninstall errors
	ErrInternalError  = errors.New("internal error")
	ErrUserRestricted = errors.New("user restricted")
	ErrAborted        = errors.New("aborted")

	// Install errors
	ErrAlreadyExists            = errors.New("already exists")
	ErrInvalidApk               = errors.New("invalid apk")
	ErrInvalidURI               = errors.New("invalid uri")
	ErrInsufficientStorage      = errors.New("insufficient storage")
	ErrDuplicatePackage         = errors.New("duplicate package")
	ErrNoSharedUser             = errors.New("no shared user")
	ErrUpdateIncompatible       = errors.New("update incompatible")
	ErrSharedUserIncompatible   = errors.New("shared user incompatible")
	ErrMissingSharedLibrary     = errors.New("missing shared library")
	ErrReplaceCouldntDelete     = errors.New("replace couldn't delete")
	ErrDexopt                   = errors.New("dexopt")
	ErrOlderSdk                 = errors.New("older sdk")
	ErrConflictingProvider      = errors.New("conflicting provider")
	ErrNewerSdk                 = errors.New("newer sdk")
	ErrTestOnly                 = errors.New("test only")
	ErrCPUAbiIncompatible       = errors.New("cpu abi incompatible")
	ErrMissingFeature           = errors.New("missing feature")
	ErrContainerError           = errors.New("combiner error")
	ErrInvalidInstallLocation   = errors.New("invalid install location")
	ErrMediaUnavailable         = errors.New("media unavailable")
	ErrVerificationTimeout      = errors.New("verification timeout")
	ErrVerificationFailure      = errors.New("verification failure")
	ErrPackageChanged           = errors.New("package changed")
	ErrUIDChanged               = errors.New("uid changed")
	ErrVersionDowngrade         = errors.New("version downgrade")
	ErrNotApk                   = errors.New("not apk")
	ErrBadManifest              = errors.New("bad manifest")
	ErrUnexpectedException      = errors.New("unexpected exception")
	ErrNoCertificates           = errors.New("no certificates")
	ErrInconsistentCertificates = errors.New("inconsistent certificates")
	ErrCertificateEncoding      = errors.New("certificate encoding")
	ErrBadPackageName           = errors.New("bad package name")
	ErrBadSharedUserID          = errors.New("bad shared user id")
	ErrManifestMalformed        = errors.New("manifest malformed")
	ErrManifestEmpty            = errors.New("manifest empty")
	ErrDuplicatePermission      = errors.New("duplicate permission")
	ErrNoMatchingAbis           = errors.New("no matching abis")

	// Uninstall errors
	ErrDevicePolicyManager = errors.New("device policy manager")
	ErrOwnerBlocked        = errors.New("owner blocked")
)

var errorVals = map[string]error{
	"FAILED_ALREADY_EXISTS":                  ErrAlreadyExists,
	"FAILED_INVALID_APK":                     ErrInvalidApk,
	"FAILED_INVALID_URI":                     ErrInvalidURI,
	"FAILED_INSUFFICIENT_STORAGE":            ErrInsufficientStorage,
	"FAILED_DUPLICATE_PACKAGE":               ErrDuplicatePackage,
	"FAILED_NO_SHARED_USER":                  ErrNoSharedUser,
	"FAILED_UPDATE_INCOMPATIBLE":             ErrUpdateIncompatible,
	"FAILED_SHARED_USER_INCOMPATIBLE":        ErrSharedUserIncompatible,
	"FAILED_MISSING_SHARED_LIBRARY":          ErrMissingSharedLibrary,
	"FAILED_REPLACE_COULDNT_DELETE":          ErrReplaceCouldntDelete,
	"FAILED_DEXOPT":                          ErrDexopt,
	"FAILED_OLDER_SDK":                       ErrOlderSdk,
	"FAILED_CONFLICTING_PROVIDER":            ErrConflictingProvider,
	"FAILED_NEWER_SDK":                       ErrNewerSdk,
	"FAILED_TEST_ONLY":                       ErrTestOnly,
	"FAILED_CPU_ABI_INCOMPATIBLE":            ErrCPUAbiIncompatible,
	"FAILED_MISSING_FEATURE":                 ErrMissingFeature,
	"FAILED_CONTAINER_ERROR":                 ErrContainerError,
	"FAILED_INVALID_INSTALL_LOCATION":        ErrInvalidInstallLocation,
	"FAILED_MEDIA_UNAVAILABLE":               ErrMediaUnavailable,
	"FAILED_VERIFICATION_TIMEOUT":            ErrVerificationTimeout,
	"FAILED_VERIFICATION_FAILURE":            ErrVerificationFailure,
	"FAILED_PACKAGE_CHANGED":                 ErrPackageChanged,
	"FAILED_UID_CHANGED":                     ErrUIDChanged,
	"FAILED_VERSION_DOWNGRADE":               ErrVersionDowngrade,
	"PARSE_FAILED_NOT_APK":                   ErrNotApk,
	"PARSE_FAILED_BAD_MANIFEST":              ErrBadManifest,
	"PARSE_FAILED_UNEXPECTED_EXCEPTION":      ErrUnexpectedException,
	"PARSE_FAILED_NO_CERTIFICATES":           ErrNoCertificates,
	"PARSE_FAILED_INCONSISTENT_CERTIFICATES": ErrInconsistentCertificates,
	"PARSE_FAILED_CERTIFICATE_ENCODING":      ErrCertificateEncoding,
	"PARSE_FAILED_BAD_PACKAGE_NAME":          ErrBadPackageName,
	"PARSE_FAILED_BAD_SHARED_USER_ID":        ErrBadSharedUserID,
	"PARSE_FAILED_MANIFEST_MALFORMED":        ErrManifestMalformed,
	"PARSE_FAILED_MANIFEST_EMPTY":            ErrManifestEmpty,
	"FAILED_INTERNAL_ERROR":                  ErrInternalError,
	"FAILED_USER_RESTRICTED":                 ErrUserRestricted,
	"FAILED_DUPLICATE_PERMISSION":            ErrDuplicatePermission,
	"FAILED_NO_MATCHING_ABIS":                ErrNoMatchingAbis,
	"FAILED_ABORTED":                         ErrAborted,
}

func parseError(s string) error {
	if err, e := errorVals[s]; e {
		return err
	}
	return fmt.Errorf("unknown error: %s", s)
}
