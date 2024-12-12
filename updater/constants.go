/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package updater

const (
	// releasePublicKeyBase64 = "RWRNqGKtBXftKTKPpBPGDMe8jHLnFQ0EdRy8Wg0apV6vTDFLAODD83G4"
	releasePublicKeyBase64 = "RWRhYmNkZWZnaPpN/6nIyyUkqTGf98Qi0hYw9vcNmi8JAUS2gdWuQbSw"
	updateServerHost       = "localhost" //"download.wireguard.com"
	updateServerPort       = 8089        //443
	updateServerUseHttps   = false
	latestVersionPath      = "/windows-client/latest.sig"
	msiPath                = "/windows-client/%s"
	msiArchPrefix          = "eclipz-%s-"
	msiSuffix              = ".msi"
)
