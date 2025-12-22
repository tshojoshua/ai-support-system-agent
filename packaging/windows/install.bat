@echo off
REM JTNT Agent - Interactive Installation Script
REM This script will guide you through the installation process

echo ========================================
echo  JTNT Agent Installation
echo ========================================
echo.

REM Check for admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator
    echo Right-click on install.bat and select "Run as administrator"
    echo.
    pause
    exit /b 1
)

echo [OK] Running as Administrator
echo.

REM Get enrollment token
set /p TOKEN="Enter enrollment token (or press Enter to skip enrollment): "

REM Get Hub URL (default)
set HUBURL=https://hub.jtnt.us
set /p CUSTOM_HUB="Hub URL [%HUBURL%]: "
if not "%CUSTOM_HUB%"=="" set HUBURL=%CUSTOM_HUB%

echo.
echo ========================================
echo  Installation Summary
echo ========================================
echo Hub URL: %HUBURL%
if "%TOKEN%"=="" (
    echo Enrollment: Manual ^(will enroll later^)
) else (
    echo Enrollment: Automatic with token
)
echo.

set /p CONFIRM="Proceed with installation? (Y/N): "
if /i not "%CONFIRM%"=="Y" (
    echo Installation cancelled.
    exit /b 0
)

echo.
echo ========================================
echo  Installing JTNT Agent...
echo ========================================

REM Find MSI file
set MSI=
for %%f in (JTNT-Agent-*.msi) do set MSI=%%f

if "%MSI%"=="" (
    echo ERROR: No MSI installer found in current directory
    echo Expected file: JTNT-Agent-*.msi
    echo.
    pause
    exit /b 1
)

echo Using installer: %MSI%
echo.

REM Install with or without enrollment
if "%TOKEN%"=="" (
    REM Install without enrollment
    echo Installing without enrollment...
    msiexec /i "%MSI%" /qb
) else (
    REM Install with enrollment
    echo Installing with enrollment...
    msiexec /i "%MSI%" /qb ENROLLMENT_TOKEN="%TOKEN%" HUB_URL="%HUBURL%"
)

REM Wait for installation
echo.
echo Waiting for installation to complete...
timeout /t 10 /nobreak >nul

REM Check if service is running
sc query JTNTAgent | find "RUNNING" >nul
if %errorLevel% equ 0 (
    echo.
    echo ========================================
    echo  Installation Successful!
    echo ========================================
    echo.
    echo Service "JTNT Agent" is now running.
    echo.

    if "%TOKEN%"=="" (
        echo To enroll the agent, run:
        echo   jtnt-agent enroll --token YOUR_TOKEN --hub %HUBURL%
        echo.
    ) else (
        echo Agent enrolled successfully.
        echo Check Hub dashboard to verify agent is online.
        echo.
    )

    echo Log files: C:\ProgramData\JTNT\Agent\logs
    echo CLI tool: jtnt-agent ^(available in PATH^)
    echo.
) else (
    echo.
    echo ========================================
    echo  Installation Complete
    echo ========================================
    echo.
    echo Service installation completed, but service is not running.
    echo Check Services (services.msc) for "JTNT Agent"
    echo.
    echo Logs: C:\ProgramData\JTNT\Agent\logs\agent.log
    echo.
)

pause
