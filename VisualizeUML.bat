@echo off
setlocal enabledelayedexpansion
title UML Visual Report Generator
color 0B

:HEADER
cls
echo.
echo  ========================================================
echo        UML Visual Report Generator
echo        Compare student UML diagrams with solution
echo  ========================================================
echo.


:: -- Step 1: Solution file --
:INPUT_SOLUTION
set "SOL_PATH="
set /p SOL_PATH="  [1/3] SOLUTION file path: "

if "!SOL_PATH!"=="" (
    echo.
    echo   [!] Path cannot be empty.
    echo.
    goto INPUT_SOLUTION
)

if not exist "!SOL_PATH!" (
    echo.
    echo   [!] File not found: "!SOL_PATH!"
    echo.
    goto INPUT_SOLUTION
)

echo   [OK] Solution verified.
echo.

:: -- Step 2: Student file --
:INPUT_STUDENT
set "STU_PATH="
set /p STU_PATH="  [2/3] STUDENT file path: "

if "!STU_PATH!"=="" (
    echo.
    echo   [!] Path cannot be empty.
    echo.
    goto INPUT_STUDENT
)

if not exist "!STU_PATH!" (
    echo.
    echo   [!] File not found: "!STU_PATH!"
    echo.
    goto INPUT_STUDENT
)

echo   [OK] Student verified.
echo.

:: -- Step 3: Output file (optional) --
set "OUT_PATH="
set /p OUT_PATH="  [3/3] Output .html name (Enter = auto): "
echo.

:: -- Confirm --
echo  --------------------------------------------------------
echo   Solution : !SOL_PATH!
echo   Student  : !STU_PATH!
if "!OUT_PATH!"=="" (
    echo   Output   : [auto-generated]
) else (
    echo   Output   : !OUT_PATH!
)
echo  --------------------------------------------------------
echo.


:: -- Run --
echo.
echo   Running visualize.exe ...
echo.

if "!OUT_PATH!"=="" (
    visualize.exe "!SOL_PATH!" "!STU_PATH!"
) else (
    visualize.exe "!SOL_PATH!" "!STU_PATH!" "!OUT_PATH!"
)

if errorlevel 1 (
    echo.
    echo   [!] visualize.exe encountered an error.
)

:: -- Loop --
echo.
echo  --------------------------------------------------------
set "YN2="
set /p YN2="  Run another comparison? (Y/N): "

if /i "!YN2!"=="Y" goto HEADER

echo.
echo   Goodbye!
echo.
pause
endlocal
