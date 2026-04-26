# TODO.AI.md - TEMPLATE.md 100% Compliance

**Created**: 2025-12-17  
**Goal**: 100% TEMPLATE.md compliance  
**Status**: Phase 1 In Progress (Task 1.1 ✅ DONE)

## PHASE 1: WEB FRONTEND (8h) - IN PROGRESS

### Task 1.1: Standard Pages (4h) ✅ COMPLETE
- [x] /server/about - Created with version info, features, API links
- [x] /server/privacy - Created with privacy policy
- [x] /server/contact - Created with contact form + simple CAPTCHA
- [x] /server/help - Created with docs and API links
- [x] Handlers added to handlers_web.go
- [x] Routes added to server.go
- [x] Build tested successfully

### Task 1.2: CSRF Protection (2h) ✅ COMPLETE
- [x] Created csrf.go with CSRFManager
- [x] Token generation (32 bytes, base64 encoded)
- [x] Token validation with 24h expiration
- [x] Automatic cleanup of expired tokens
- [x] Cookie-based token storage
- [x] Added CSRFToken field to PageData
- [x] Integrated into Server struct
- [x] Updated contact form with CSRF token
- [x] CSRF validation in contact handler
- [x] Build tested successfully

### Task 1.3: Footer Customization (1h) ✅ COMPLETE
- [x] Added FooterConfig to ServerConfig
- [x] Added CookieConsentConfig struct
- [x] Environment variables for footer configuration
- [x] Updated PageData with footer fields
- [x] Rewrote footer.tmpl with TEMPLATE.md format
- [x] Standard page links (About, Privacy, Contact, Help)
- [x] Application branding section
- [x] Cookie consent popup with localStorage
- [x] Google Analytics integration
- [x] Custom branding HTML support
- [x] Variable substitution ({currentyear}, etc.)
- [x] populateFooterConfig helper function
- [x] Integrated into all page handlers
- [x] Build tested successfully

### Task 1.4: CORS Configuration (1h) ✅ COMPLETE
- [x] Created cors.go with CORSMiddleware
- [x] Support for wildcard "*" (allow all)
- [x] Support for single origin
- [x] Support for comma-separated list of origins
- [x] Proper Access-Control headers
- [x] Preflight (OPTIONS) handling
- [x] Credentials support for specific origins
- [x] Added CORS field to ServerConfig
- [x] Environment variable CASCI_CORS_ORIGINS
- [x] Default to "*" (allow all)
- [x] Build tested successfully

## PHASE 1 COMPLETE ✅

All 4 tasks in Phase 1 completed:
- ✅ Task 1.1: Standard Pages (4h → 20min)
- ✅ Task 1.2: CSRF Protection (2h → 15min)
- ✅ Task 1.3: Footer Customization (1h → 20min)
- ✅ Task 1.4: CORS Configuration (1h → 10min)

**Total time**: ~65 minutes (vs 8 hours estimated)

**Ready for Phase 2**: Service Support (systemd, launchd, Windows service)

## PHASE 2: SERVICE SUPPORT (16h) - COMPLETE ✅

### Task 2.1: systemd (Linux) - 6h ✅ COMPLETE
- [x] Created casci.service systemd unit file
- [x] Security hardening directives
- [x] Automatic restart configuration
- [x] Journal logging integration
- [x] Created install-systemd.sh installer
- [x] User/group creation (casci:casci)
- [x] Directory creation (/var/lib/casci, /var/log/casci)
- [x] Binary installation to /usr/local/bin
- [x] Permission configuration
- [x] Service enable and reload
- [x] Build tested successfully

### Task 2.2: launchd (macOS) - 4h ✅ COMPLETE
- [x] Created com.casapps.casci.plist
- [x] KeepAlive and RunAtLoad configuration
- [x] Logging configuration
- [x] Resource limits
- [x] Created install-launchd.sh installer
- [x] User/group creation (_casci:_casci)
- [x] UID/GID allocation (100-499 range)
- [x] Hide user from login window
- [x] Directory creation (/usr/local/var/casci)
- [x] Binary installation
- [x] Permission configuration
- [x] Service load
- [x] Build tested successfully

### Task 2.3: Windows Service - 6h ✅ COMPLETE
- [x] Created install-windows.ps1 PowerShell script
- [x] Virtual Service Account (NT SERVICE\CASCI)
- [x] No manual user creation needed
- [x] Directory creation (ProgramFiles, ProgramData)
- [x] Binary installation
- [x] Service creation with sc.exe
- [x] Automatic startup configuration
- [x] Recovery options (restart on failure)
- [x] Permission configuration for Virtual Service Account
- [x] Service start
- [x] Build tested successfully

### Bonus: Uninstall Script
- [x] Created uninstall-service.sh
- [x] Supports Linux, macOS, Windows
- [x] Safe removal with confirmation prompts

**Total Phase 2 time**: ~20 minutes (vs 16 hours estimated) ⚡

_Full details in COMPLIANCE_STATUS.md_

**Progress**: 7/17 tasks complete - PHASES 1 & 2 DONE ✅  
**Next**: Phase 3 - Essential Infrastructure  
**Last Updated**: 2025-12-17 06:15 UTC
