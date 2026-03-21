@echo off
setlocal enabledelayedexpansion
set EXE=CompareUML_CLI.exe

if not exist %EXE% (
    echo Error: %EXE% not found! Please build it first.
    pause
    exit /b 1
)

if "%~1"=="" (
    echo ==================================================
    echo       Interactive UML Compare CLI Launcher
    echo ==================================================
    
    :get_sol
    set /p "SOL_FILE=Nhap ten file DAP AN (Solution) [.drawio]: "
    if not exist "!SOL_FILE!" (
        echo [ERROR] Khong tim thay file: !SOL_FILE!
        goto get_sol
    )

    :get_stu
    set /p "STU_FILE=Nhap ten file BAI LAM (Student) [.drawio]: "
    if not exist "!STU_FILE!" (
        echo [ERROR] Khong tim thay file: !STU_FILE!
        goto get_stu
    )

    echo.
    echo Dang so sanh: !SOL_FILE! vs !STU_FILE!
    echo --------------------------------------------------
    %EXE% "!SOL_FILE!" "!STU_FILE!"
    pause
) else (
    %EXE% %*
)
goto get_sol
endlocal
