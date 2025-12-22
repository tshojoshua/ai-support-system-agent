@echo off
REM JTNT Agent - Silent Installation Script for Deployment
REM Usage: install-silent.bat <enrollment-token> [hub-url]

if "%1"=="" (
    echo ERROR: Enrollment token required
    echo Usage: install-silent.bat ^<enrollment-token^> [hub-url]
    exit /b 1
)

set TOKEN=%1
set HUBURL=%2
if "%HUBURL%"=="" set HUBURL=https://hub.jtnt.us

REM Find MSI file
set MSI=
for %%f in (JTNT-Agent-*.msi) do set MSI=%%f

if "%MSI%"=="" (
    echo ERROR: No MSI installer found
    exit /b 1
)

REM Silent install with enrollment
msiexec /i "%MSI%" /qn /l*v install.log ENROLLMENT_TOKEN="%TOKEN%" HUB_URL="%HUBURL%"

exit /b %errorlevel%
