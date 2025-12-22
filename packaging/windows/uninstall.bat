@echo off
REM JTNT Agent - Uninstallation Script

echo ========================================
echo  JTNT Agent Uninstallation
echo ========================================
echo.

REM Check for admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator
    echo.
    pause
    exit /b 1
)

echo [OK] Running as Administrator
echo.

set /p CONFIRM="Are you sure you want to uninstall JTNT Agent? (Y/N): "
if /i not "%CONFIRM%"=="Y" (
    echo Uninstallation cancelled.
    exit /b 0
)

echo.
echo Uninstalling JTNT Agent...

REM Uninstall using product name
wmic product where "name='JTNT Agent'" call uninstall /nointeractive

echo.
echo Uninstallation complete.
echo.
echo Note: Configuration and logs remain in C:\ProgramData\JTNT\Agent
echo Delete manually if you want to remove all data.
echo.

pause
