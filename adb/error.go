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

func parseError(s string) error {
	switch s {
	case "FAILED_ALREADY_EXISTS":
		return ErrAlreadyExists
	case "FAILED_INVALID_APK":
		return ErrInvalidApk
	case "FAILED_INVALID_URI":
		return ErrInvalidURI
	case "FAILED_INSUFFICIENT_STORAGE":
		return ErrInsufficientStorage
	case "FAILED_DUPLICATE_PACKAGE":
		return ErrDuplicatePackage
	case "FAILED_NO_SHARED_USER":
		return ErrNoSharedUser
	case "FAILED_UPDATE_INCOMPATIBLE":
		return ErrUpdateIncompatible
	case "FAILED_SHARED_USER_INCOMPATIBLE":
		return ErrSharedUserIncompatible
	case "FAILED_MISSING_SHARED_LIBRARY":
		return ErrMissingSharedLibrary
	case "FAILED_REPLACE_COULDNT_DELETE":
		return ErrReplaceCouldntDelete
	case "FAILED_DEXOPT":
		return ErrDexopt
	case "FAILED_OLDER_SDK":
		return ErrOlderSdk
	case "FAILED_CONFLICTING_PROVIDER":
		return ErrConflictingProvider
	case "FAILED_NEWER_SDK":
		return ErrNewerSdk
	case "FAILED_TEST_ONLY":
		return ErrTestOnly
	case "FAILED_CPU_ABI_INCOMPATIBLE":
		return ErrCPUAbiIncompatible
	case "FAILED_MISSING_FEATURE":
		return ErrMissingFeature
	case "FAILED_CONTAINER_ERROR":
		return ErrContainerError
	case "FAILED_INVALID_INSTALL_LOCATION":
		return ErrInvalidInstallLocation
	case "FAILED_MEDIA_UNAVAILABLE":
		return ErrMediaUnavailable
	case "FAILED_VERIFICATION_TIMEOUT":
		return ErrVerificationTimeout
	case "FAILED_VERIFICATION_FAILURE":
		return ErrVerificationFailure
	case "FAILED_PACKAGE_CHANGED":
		return ErrPackageChanged
	case "FAILED_UID_CHANGED":
		return ErrUIDChanged
	case "FAILED_VERSION_DOWNGRADE":
		return ErrVersionDowngrade
	case "PARSE_FAILED_NOT_APK":
		return ErrNotApk
	case "PARSE_FAILED_BAD_MANIFEST":
		return ErrBadManifest
	case "PARSE_FAILED_UNEXPECTED_EXCEPTION":
		return ErrUnexpectedException
	case "PARSE_FAILED_NO_CERTIFICATES":
		return ErrNoCertificates
	case "PARSE_FAILED_INCONSISTENT_CERTIFICATES":
		return ErrInconsistentCertificates
	case "PARSE_FAILED_CERTIFICATE_ENCODING":
		return ErrCertificateEncoding
	case "PARSE_FAILED_BAD_PACKAGE_NAME":
		return ErrBadPackageName
	case "PARSE_FAILED_BAD_SHARED_USER_ID":
		return ErrBadSharedUserID
	case "PARSE_FAILED_MANIFEST_MALFORMED":
		return ErrManifestMalformed
	case "PARSE_FAILED_MANIFEST_EMPTY":
		return ErrManifestEmpty
	case "FAILED_INTERNAL_ERROR":
		return ErrInternalError
	case "FAILED_USER_RESTRICTED":
		return ErrUserRestricted
	case "FAILED_DUPLICATE_PERMISSION":
		return ErrDuplicatePermission
	case "FAILED_NO_MATCHING_ABIS":
		return ErrNoMatchingAbis
	case "FAILED_ABORTED":
		return ErrAborted
	}
	return fmt.Errorf("unknown error: %s", s)
}
